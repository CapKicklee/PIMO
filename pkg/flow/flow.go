// Copyright (C) 2022 CGI France
//
// This file is part of PIMO.
//
// PIMO is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// PIMO is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with PIMO.  If not, see <http://www.gnu.org/licenses/>.

package flow

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/cgi-fr/pimo/pkg/model"
)

type subgraph struct {
	name    string
	masks   []edge
	removed bool
	added   bool
}

type edge struct {
	source      string
	destination string
	mask        string
	param       string
	key         string
}

func Export(masking model.Definition) (string, error) {
	maskingDef := masking.Masking
	if len(maskingDef) == 0 {
		return "", nil
	}
	res := `flowchart LR
    `
	variables := make(map[string]subgraph)
	maskOrder := make([]string, 0, 10)
	for i := 0; i < len(maskingDef); i++ {
		_, ok := variables[maskingDef[i].Selector.Jsonpath]
		if !ok {
			maskOrder = append(maskOrder, maskingDef[i].Selector.Jsonpath)
		}
		if maskingDef[i].Masks != nil {
			for _, v := range maskingDef[i].Masks {
				exportMask(maskingDef[i], v, variables)
			}
		}
		exportMask(maskingDef[i], maskingDef[i].Mask, variables)
	}
	res += printSubgraphs(variables, maskOrder)
	return res, nil
}

func exportMask(masking model.Masking, mask model.MaskType, variables map[string]subgraph) {
	maskSubgraph := subgraph{
		name:    masking.Selector.Jsonpath + "_sg",
		removed: false,
		added:   false,
		masks:   make([]edge, 0, 10),
	}
	edgeToAdd := edge{}
	edgeToAdd.key = masking.Selector.Jsonpath
	edgeToAdd.source = masking.Selector.Jsonpath
	edgeToAdd.destination = masking.Selector.Jsonpath + "_1"
	if elem, ok := variables[masking.Selector.Jsonpath]; ok {
		maskSubgraph = elem
		edgeToAdd = checkSourceAndDestination(edgeToAdd, len(elem.masks), maskSubgraph, masking)
	}

	if mask.Add != nil {
		edgeToAdd.mask = "Add"
		edgeToAdd.param = mask.Add.(string)
		maskSubgraph = exportAdd(maskSubgraph, edgeToAdd, mask, masking, variables)
		maskSubgraph.added = true
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.AddTransient != nil {
		edgeToAdd.mask = "AddTransient"
		edgeToAdd.param = mask.AddTransient.(string)
		maskSubgraph = exportAddTransient(maskSubgraph, edgeToAdd, mask, masking, variables)
		maskSubgraph.added = true
		maskSubgraph.removed = true
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.Constant != nil {
		edgeToAdd.mask = "Constant"
		edgeToAdd.param = mask.Constant.(string)
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.RandomChoice != nil {
		edgeToAdd.mask = "RandomChoice"
		edgeToAdd.param = flattenChoices(mask)
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.RandomChoiceInURI != "" {
		edgeToAdd.mask = "RandomChoiceInURI"
		edgeToAdd.param = mask.RandomChoiceInURI
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.Command != "" {
		edgeToAdd.mask = "Command"
		edgeToAdd.param = mask.Command
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.RandomInt != (model.RandIntType{}) {
		edgeToAdd.mask = "RandomInt"
		edgeToAdd.param = "Min: " + strconv.Itoa(mask.RandomInt.Min) + ", Max: " + strconv.Itoa(mask.RandomInt.Max)
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if len(mask.WeightedChoice) > 0 {
		edgeToAdd.mask = "WeightedChoice"
		edgeToAdd.param = flattenWeightedChoices(mask)
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.Regex != "" {
		edgeToAdd.mask = "Regex"
		edgeToAdd.param = mask.Regex
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.Hash != nil {
		edgeToAdd.mask = "Hash"
		edgeToAdd.param = flattenHash(mask)
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.HashInURI != "" {
		edgeToAdd.mask = "HashInURI"
		edgeToAdd.param = mask.HashInURI
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.RandDate != (model.RandDateType{}) {
		edgeToAdd.mask = "RandDate"
		edgeToAdd.param = "DateMin: " + mask.RandDate.DateMin.String() + ", DateMax: " + mask.RandDate.DateMax.String()
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.Incremental != (model.IncrementalType{}) {
		edgeToAdd.mask = "Incremental"
		edgeToAdd.param = "Start: " + strconv.Itoa(mask.Incremental.Start) + ", Increment: " + strconv.Itoa(mask.Incremental.Increment)
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.Replacement != "" {
		edgeToAdd.mask = "Replacement"
		edgeToAdd.param = mask.Replacement
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.Template != "" {
		edgeToAdd.mask = "Template"
		edgeToAdd.param = mask.Template
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		maskSubgraph = unescapeTemplateValues(mask.Template, "Template", masking.Selector.Jsonpath, variables, maskSubgraph)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.TemplateEach != (model.TemplateEachType{}) {
		edgeToAdd.mask = "TemplateEach"
		edgeToAdd.param = "Item: " + mask.TemplateEach.Item + ", Index: " + mask.TemplateEach.Index + ", Template: " + mask.TemplateEach.Template
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		maskSubgraph = unescapeTemplateValues(mask.TemplateEach.Template, "TemplateEach", masking.Selector.Jsonpath, variables, maskSubgraph)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.Duration != "" {
		edgeToAdd.mask = "Duration"
		edgeToAdd.param = mask.Duration
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.Remove {
		edgeToAdd.mask = "Remove"
		edgeToAdd.param = strconv.FormatBool(true)
		maskSubgraph.removed = true
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.RangeMask != 0 {
		edgeToAdd.mask = "RangeMask"
		edgeToAdd.param = strconv.Itoa(mask.RangeMask)
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.RandomDuration != (model.RandomDurationType{}) {
		edgeToAdd.mask = "RandomDuration"
		edgeToAdd.param = "Min: " + mask.RandomDuration.Min + ", Max: " + mask.RandomDuration.Max
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.FluxURI != "" {
		edgeToAdd.mask = "FluxURI"
		edgeToAdd.param = mask.FluxURI
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.RandomDecimal != (model.RandomDecimalType{}) {
		edgeToAdd.mask = "RandomDecimal"
		min := strconv.FormatFloat(mask.RandomDecimal.Min, 'E', mask.RandomDecimal.Precision, 64)
		max := strconv.FormatFloat(mask.RandomDecimal.Max, 'E', mask.RandomDecimal.Precision, 64)
		precision := strconv.Itoa(mask.RandomDecimal.Precision)
		edgeToAdd.param = "Min: " + min + ", Max: " + max + ", Precision: " + precision
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.DateParser != (model.DateParserType{}) {
		edgeToAdd.mask = "DateParser"
		edgeToAdd.param = "InputFormat: " + mask.DateParser.InputFormat + ", OutputFormat: " + mask.DateParser.OutputFormat
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.FromCache != "" {
		edgeToAdd.mask = "FromCache"
		edgeToAdd.param = mask.FromCache
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.FF1 != (model.FF1Type{}) {
		edgeToAdd.mask = "FF1"
		edgeToAdd.param = "KeyFromEnv: " + mask.FF1.KeyFromEnv + ", TweakField: " + mask.FF1.TweakField + ", Radix: " + strconv.FormatUint(uint64(mask.FF1.Radix), 10) + ", Decrypt: " + strconv.FormatBool(mask.FF1.Decrypt)
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.Pipe.Masking != nil {
		edgeToAdd.mask = "Pipe"
		edgeToAdd.param = "DefinitionFile: " + mask.Pipe.DefinitionFile + ", InjectParent: " + mask.Pipe.InjectParent + ", InjectRoot: " + mask.Pipe.InjectRoot
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.FromJSON != "" {
		edgeToAdd.mask = "FromJSON"
		edgeToAdd.param = mask.FromJSON
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
	if mask.Luhn != nil {
		edgeToAdd.mask = "Luhn"
		edgeToAdd.param = mask.Luhn.Universe
		maskSubgraph.masks = append(maskSubgraph.masks, edgeToAdd)
		variables[masking.Selector.Jsonpath] = maskSubgraph
	}
}

func checkSourceAndDestination(edgeToAdd edge, maskCount int, maskSubgraph subgraph, masking model.Masking) edge {
	if maskCount > 0 {
		edgeToAdd.source = maskSubgraph.masks[len(maskSubgraph.masks)-1].destination
		edgeToAdd.destination = masking.Selector.Jsonpath + "_" + strconv.Itoa(len(maskSubgraph.masks)+1)
	}
	return edgeToAdd
}

func exportAdd(maskSubgraph subgraph, addEdge edge, mask model.MaskType, masking model.Masking, variables map[string]subgraph) subgraph {
	if mask.Add.(string) != "" {
		maskSubgraph.masks = append(maskSubgraph.masks, addEdge)
	}
	if strings.Contains(mask.Add.(string), "{{") {
		maskSubgraph = unescapeTemplateValues(mask.Add.(string), "Add", masking.Selector.Jsonpath, variables, maskSubgraph)
	}
	maskSubgraph.added = true
	return maskSubgraph
}

func exportAddTransient(maskSubgraph subgraph, addEdge edge, mask model.MaskType, masking model.Masking, variables map[string]subgraph) subgraph {
	if mask.AddTransient.(string) != "" {
		maskSubgraph.masks = append(maskSubgraph.masks, addEdge)
	}
	if strings.Contains(mask.AddTransient.(string), "{{") {
		maskSubgraph = unescapeTemplateValues(mask.AddTransient.(string), "AddTransient", masking.Selector.Jsonpath, variables, maskSubgraph)
	}
	maskSubgraph.added = true
	return maskSubgraph
}

func flattenChoices(mask model.MaskType) string {
	choices := make([]string, len(mask.RandomChoice))
	for i, v := range mask.RandomChoice {
		choices[i] = v.(string)
	}
	return strings.Join(choices, ",")
}

func flattenWeightedChoices(mask model.MaskType) string {
	choices := make([]string, len(mask.WeightedChoice))
	for i, v := range mask.WeightedChoice {
		weight := uint64(v.Weight)
		weightStr := strconv.FormatUint(weight, 10)
		choices[i] = v.Choice.(string) + " @ " + weightStr
	}
	return strings.Join(choices, ",")
}

func flattenHash(mask model.MaskType) string {
	choices := make([]string, len(mask.Hash))
	for i, v := range mask.Hash {
		choices[i] = v.(string)
	}
	return strings.Join(choices, ",")
}

func unescapeTemplateValues(templateValue, mask, jsonpath string, variables map[string]subgraph, maskSubgraph subgraph) subgraph {
	regex := regexp.MustCompile(`(?:{{\.)([0-z]+)(?:}})`)
	splittedTemplate := regex.FindAllString(templateValue, -1)
	edges := make([]edge, 0, 10)
	copyarray := make([]edge, 0, 10)
	copy(copyarray, variables[jsonpath].masks)
	jsonpathMaskCount := len(copyarray)
	for i := range splittedTemplate {
		templateEdge := edge{
			mask:  mask,
			param: templateValue,
		}

		value := splittedTemplate[i][3 : len(splittedTemplate[i])-2]

		templateEdge.key = value

		maskNumber := len(variables[value].masks)

		if maskNumber == 0 {
			templateEdge.source = value
		} else {
			templateEdge.source = value + "_" + strconv.Itoa(maskNumber)
		}

		templateEdge.destination = jsonpath + "_" + strconv.Itoa(jsonpathMaskCount+1)
		edges = append(edges, templateEdge)
	}
	maskSubgraph.masks = append(maskSubgraph.masks, edges...)
	return maskSubgraph
}

func printSubgraphs(variables map[string]subgraph, maskOrder []string) string {
	subgraphText := ""
	for _, key := range maskOrder {
		if variables[key].added {
			subgraphText += "!add[/Add/] --> " + key + "\n    "
		} else {
			subgraphText += "!input[(input)] --> " + key + "\n    "
		}

		count := len(variables[key].masks)
		if count > 0 {
			subgraphText += "subgraph " + variables[key].name
			for j := range variables[key].masks {
				subgraphText = printMask(subgraphText, variables[key].masks[j], variables)
			}
			subgraphText += "\n    end\n    "
			subgraphText += variables[key].masks[count-1].destination
		} else {
			subgraphText += key
		}
		if variables[key].removed {
			subgraphText += " --> !remove[\\Remove\\]\n    "
		} else {
			subgraphText += " --> !output>Output]\n    "
		}
	}
	return strings.TrimSpace(subgraphText)
}

func printMask(subgraphText string, mask edge, variables map[string]subgraph) string {
	_, ok := variables[mask.key]
	if !ok {
		subgraphText += "\n        !input[(input)] --> " + mask.key
	}
	subgraphText += "\n        " + mask.source + " -->|\"" + mask.mask + "(" + mask.param + ")\"| " + mask.destination
	return subgraphText
}