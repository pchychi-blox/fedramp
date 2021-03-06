package template

import (
	"errors"
	"fmt"
	"github.com/gocomply/fedramp/pkg/docx_helper"
	"github.com/gocomply/fedramp/pkg/fedramp"
	"github.com/gocomply/fedramp/pkg/templater/template/checkbox"
	"github.com/jbowtie/gokogiri/xml"
	"regexp"
	"strings"
)

// ControlSummaryInformation represents single table labeled "Control Summary Information"
type ControlSummaryInformation struct {
	table xml.Node
}

func (t *Template) ControlSummaryInformations() ([]ControlSummaryInformation, error) {
	tables, err := t.querySummaryTables()
	if err != nil {
		return nil, err
	}

	var result []ControlSummaryInformation
	for _, table := range tables {
		result = append(result, ControlSummaryInformation{
			table: table,
		})
	}
	return result, nil
}

func (t *Template) querySummaryTables() ([]xml.Node, error) {
	return t.xmlDoc.Search(
		"//w:tbl[contains(normalize-space(.), 'Control Summary') or contains(normalize-space(.), 'Control Enhancement Summary')]",
	)
}

func (csi *ControlSummaryInformation) ControlName() (name string, err error) {
	content, err := csi.queryTableHeader()
	if err != nil {
		return
	}

	if content == "CM2 (7)Control Summary Information" {
		// Workaround typo in the 8/28/2018 version of FedRAMP-SSP-High-Baseline-Template.docx
		content = "CM-2 (7)Control Summary Information"
	}

	name, err = parseControlId(content)
	return
}

func (csi *ControlSummaryInformation) queryTableHeader() (string, error) {
	content, err := docx_helper.ParseTableHeaderContent(csi.table)
	if err != nil {
		return "", fmt.Errorf("Control Name in 'Control Summary Information' table not found: %s", err)

	}
	return content, err
}

// Represents "Responsible Role" row in the Control Summary Table (usually first row)
type ResponsibleRole struct {
	node xml.Node
}

func (csi *ControlSummaryInformation) ResponsibleRole() (*ResponsibleRole, error) {
	nodes, err := csi.table.Search(".//w:tc[starts-with(normalize-space(.), 'Responsible Role')]")
	if err != nil {
		return nil, err
	}
	if len(nodes) != 1 {
		name, err := csi.ControlName()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Could not find Responsible Role cell in Control Summary Information Table of %s", name)
	}

	return &ResponsibleRole{node: nodes[0]}, nil
}

func (rr *ResponsibleRole) SetValue(roleName string) error {
	textNodes, err := rr.node.Search(".//w:t")
	if err != nil || len(textNodes) < 1 {
		return errors.New("Cannot find any child text nodes when processing Responsible Role column")
	}
	return textNodes[0].SetContent(fmt.Sprintf("Responsible Role: %s", roleName))
}

type ParameterRow struct {
	node xml.Node
}

func (csi *ControlSummaryInformation) ParameterRows() ([]ParameterRow, error) {
	rows, err := csi.table.Search(".//w:tc[starts-with(normalize-space(.), 'Parameter')]")
	if err != nil {
		return nil, err
	}
	var result []ParameterRow
	for _, row := range rows {
		result = append(result, ParameterRow{
			node: row,
		})
	}
	return result, nil
}

func (pr *ParameterRow) ParamId() (string, error) {
	nodes, err := pr.node.Search(".//w:t[starts-with(normalize-space(.), 'Parameter')]")
	if err != nil {
		return "", err
	}
	if len(nodes) != 1 {
		return "", fmt.Errorf("Could not find Parameter text field in Control Summary table")
	}
	txt, err := docx_helper.ConcatTextNodes(pr.node)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`^Parameter\s+([^:]*):*\s*(Not applicable)?$`)
	match := re.FindStringSubmatch(txt)
	if len(match) == 0 {
		return "", fmt.Errorf("Could not locate parameter ID in text: '%s'", txt)
	}
	id := match[1]

	return id, nil
}

func (pr *ParameterRow) ControlId() (string, error) {
	paramId, err := pr.ParamId()
	if err != nil {
		return "", err
	}
	paramId = strings.ToLower(paramId)

	re := regexp.MustCompile(`([a-z][a-z]-[0-9]+)`)
	match := re.FindStringSubmatch(paramId)
	if len(match) == 0 {
		return "", fmt.Errorf("Could not translate '%s' to NIST-800-53 control id", paramId)
	}
	return match[1], nil
}

func (pr *ParameterRow) SetValue(paramText string) error {
	textNodes, err := pr.node.Search(".//w:t")
	if err != nil || len(textNodes) < 1 {
		return errors.New("Cannot find any child text nodes when processing Parametr row")
	}
	tn := textNodes[len(textNodes)-1]
	return tn.SetContent(fmt.Sprintf("%s %s", tn.Content(), paramText))
}

type ImplementationStatus struct {
	node     xml.Node
	statuses map[fedramp.ImplementationStatus]*checkbox.CheckBox
}

func (csi *ControlSummaryInformation) ImplementationStatus() (*ImplementationStatus, error) {
	rows, err := csi.table.Search(".//w:tc[starts-with(normalize-space(.), 'Implementation Status')]")
	if err != nil {
		return nil, err
	}
	if len(rows) != 1 {
		name, err := csi.ControlName()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Could not find 'Implementation Status' cell in Control Summary Information Table of %s", name)
	}
	return parseImplementationStatus(rows[0])
}

func parseImplementationStatus(node xml.Node) (is *ImplementationStatus, err error) {
	paragraphs, err := node.Search(".//w:p")
	if err != nil {
		return
	}
	statuses := map[fedramp.ImplementationStatus]*checkbox.CheckBox{}
	for _, paragraph := range paragraphs {
		cb, err := checkbox.Parse(paragraph)
		if err != nil {
			if _, ok := err.(*checkbox.NotFound); ok {
				continue
			}
			return nil, err
		}
		cbStatus := fedramp.StatusFromDocx(cb.Text())
		statuses[cbStatus] = cb
	}

	return &ImplementationStatus{node: node, statuses: statuses}, nil
}

func (is *ImplementationStatus) SetValue(newStatus fedramp.ImplementationStatus) error {
	cb, found := is.statuses[newStatus]
	if found {
		cb.SetChecked()
	}
	return nil
}
