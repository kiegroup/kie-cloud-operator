import React from "react";

import {
  FormGroup,
  FormSelectOption,
  FormSelect,
  Tooltip
} from "@patternfly/react-core";

import * as utils from "../../../common/CommonUtils";

import JSONPATH from "jsonpath";

export class DropdownField {
  constructor(props) {
    this.props = props;
    this.errMsg = "";
    this.isValid = true;
  }

  getJsx() {
    var options = [{ value: "", label: "" }];

    const tmpJsonPath = utils.getJsonSchemaPathForJsonPath(
      this.props.fieldDef.originalJsonPath
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
    if (this.props.fieldDef.value === undefined) {
      this.props.fieldDef.value = "";
    }
    if (
      this.props.fieldDef.required === true &&
      (this.props.fieldDef.value === undefined ||
        this.props.fieldDef.value === "")
    ) {
      this.errMsg = this.props.fieldDef.label + " is required.";
      this.isValid = false;
    } else {
      this.errMsg = "";
      this.isValid = true;
    }
    this.props.fieldDef.errMsg = this.errMsg;
    return (
      <FormGroup
        label={this.props.fieldDef.label}
        id={this.props.ids.fieldGroupId}
        key={this.props.ids.fieldGroupKey}
        helperTextInvalid={this.errMsg}
        fieldId={this.props.ids.fieldId}
        isValid={this.isValid}
        isRequired={this.props.fieldDef.required}
      >
        <Tooltip
          position="left"
          content={<div>{this.props.fieldDef.description}</div>}
          enableFlip={true}
          style={{
            display:
              this.props.fieldDef.description !== undefined &&
              this.props.fieldDef.description !== ""
                ? "block"
                : "none"
          }}
        >
          <FormSelect
            id={this.props.ids.fieldId}
            key={this.props.ids.fieldKey}
            name={this.props.fieldDef.label}
            jsonpath={this.props.fieldDef.jsonPath}
            onChange={this.onSelect}
            value={this.props.fieldDef.value}
          >
            {options.map((option, index) => (
              <FormSelectOption
                key={this.props.ids.fieldKey + index}
                value={option.value}
                label={option.label}
              />
            ))}
          </FormSelect>
        </Tooltip>
      </FormGroup>
    );
  }

  onSelect = (_, event) => {
    let value = event.target.value;
    this.props.fieldDef.value = value;

    if (this.props.fieldDef.required === true && value === "") {
      this.errMsg = this.props.fieldDef.label + " is required.";
      this.isValid = false;
    } else {
      this.errMsg = "";
      this.isValid = true;
    }
    this.props.fieldDef.errMsg = this.errMsg;
    this.props.page.loadPageChildren();
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
