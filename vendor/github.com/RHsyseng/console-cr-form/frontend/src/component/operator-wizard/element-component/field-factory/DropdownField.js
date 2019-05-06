import React from "react";

import {
  FormGroup,
  FormSelectOption,
  FormSelect
} from "@patternfly/react-core";

import * as utils from "../../../common/CommonUtils";

import JSONPATH from "jsonpath";

export class DropdownField {
  constructor(props) {
    this.props = props;
  }

  getJsx() {
    var options = [];

    const tmpJsonPath = utils.getJsonSchemaPathForJsonPath(
      this.props.fieldDef.jsonPath
    );
    const optionValues = this.findValueFromSchema(tmpJsonPath + ".enum");
    if (optionValues !== undefined) {
      optionValues.forEach(option => {
        const oneOption = {
          label: option,
          value: option
        };
        options.push(oneOption);
      });
    }
    //const helpText = this.findValueFromSchema(tmpJsonPath + ".description");
    if (this.props.fieldDef.required === true) {
      return (
        <FormGroup
          label={this.props.fieldDef.label}
          fieldId={this.props.ids.fieldGroupId}
          key={this.props.ids.fieldGroupKey}
          // helperText={helpText}
          isRequired
        >
          <FormSelect
            id={this.props.ids.fieldId}
            key={this.props.ids.fieldKey}
            name={this.props.fieldDef.label}
            jsonpath={this.props.fieldDef.jsonPath}
            onChange={this.onSelect}
            defaultValue={this.props.fieldDef.value}
            isRequired
          >
            {options.map((option, index) => (
              <FormSelectOption
                isDisabled={option.disabled}
                id={this.props.ids.fieldId + index}
                key={this.props.ids.fieldKey + index}
                value={option.value}
                label={option.label}
              />
            ))}
          </FormSelect>
        </FormGroup>
      );
    } else {
      return (
        <FormGroup
          label={this.props.fieldDef.label}
          fieldId={this.props.ids.fieldGroupId}
          key={this.props.ids.fieldGroupKey}
          // helperText={helpText}
        >
          <FormSelect
            id={this.props.ids.fieldId}
            key={this.props.ids.fieldKey}
            name={this.props.fieldDef.label}
            jsonpath={this.props.fieldDef.jsonPath}
            onChange={this.onSelect}
            defaultValue={this.props.fieldDef.value}
          >
            {options.map((option, index) => (
              <FormSelectOption
                isDisabled={option.disabled}
                id={this.props.ids.fieldId + index}
                key={this.props.ids.fieldKey + index}
                value={option.value}
                label={option.label}
              />
            ))}
          </FormSelect>
        </FormGroup>
      );
    }
  }

  onSelect = value => {
    this.props.fieldDef.value = value;
  };
  findValueFromSchema(jsonPath) {
    try {
      var queryResults = JSONPATH.query(this.props.jsonSchema, jsonPath);

      if (Array.isArray(queryResults) && queryResults.length > 0) {
        return queryResults[0];
      }
    } catch (error) {
      console.debug("Failed to find a value from schema", error);
    }
    return [];
  }
}
