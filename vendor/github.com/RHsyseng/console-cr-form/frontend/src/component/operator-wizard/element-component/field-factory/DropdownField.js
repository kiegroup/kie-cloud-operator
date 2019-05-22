import React from "react";

import {
  FormGroup,
  FormSelectOption,
  FormSelect
} from "@patternfly/react-core";

import FieldFactory from "./FieldFactory";
import JSONPATH from "jsonpath";

export class DropdownField {
  constructor(props) {
    this.props = props;
    if (
      props.fieldDef.value === undefined &&
      props.fieldDef.default !== undefined
    ) {
      this.props.fieldDef.value = props.fieldDef.default;
    }
    this.errMsg = "";
    this.isValid = true;
    this.addChildren = this.addChildren.bind(this);
    this.reBuildChildren = this.reBuildChildren.bind(this);
  }

  getJsonSchemaPathForJsonPath(jsonPath) {
    if (jsonPath !== undefined && jsonPath !== "") {
      jsonPath = jsonPath.slice(2, jsonPath.length);
      jsonPath = jsonPath.replace(/\./g, ".properties.");
      jsonPath = "$.." + jsonPath;
    }
    return jsonPath;
  }

  getJsx() {
    let options = [{ value: "", label: "Select here" }];

    if (this.props.fieldDef.options) {
      options = this.props.fieldDef.options;
    } else {
      const tmpJsonPath = this.getJsonSchemaPathForJsonPath(
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

    var jsxArray = [];
    jsxArray.push(
      <FormGroup
        label={this.props.fieldDef.label}
        id={this.props.ids.fieldGroupId}
        key={this.props.ids.fieldGroupKey}
        helperTextInvalid={this.errMsg}
        helperText={this.props.fieldDef.description}
        fieldId={this.props.ids.fieldId}
        isValid={this.isValid}
        isRequired={this.props.fieldDef.required}
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
      </FormGroup>
    );
    jsxArray.push(this.addChildren());
    return jsxArray;
  }

  addChildren() {
    var elements = [];

    if (this.props.fieldDef.fields) {
      this.props.fieldDef.fields.forEach((subfield, i) => {
        var parentjsonpath = this.props.fieldDef.jsonPath;
        if (parentjsonpath !== undefined && parentjsonpath !== "") {
          parentjsonpath = parentjsonpath.slice(
            0,
            parentjsonpath.lastIndexOf(".")
          );
          var res = "";
          if (parentjsonpath.length < subfield.jsonPath.length) {
            res = subfield.jsonPath.substring(
              parentjsonpath.length,
              subfield.jsonPath.length
            );
            subfield.jsonPath = parentjsonpath.concat(res);
          }
        }
        if (subfield.type != "object") {
          let oneComponent = FieldFactory.newInstance(
            subfield,
            i,
            this.props.pageNumber,
            this.props.jsonSchema,
            this.props.page
          );
          elements.push(oneComponent.getJsx());
        } else {
          if (subfield.label === this.props.fieldDef.value) {
            //when the drop down value matches field group
            subfield.visible = true;
          }
          let oneComponent = FieldFactory.newInstance(
            subfield,
            i,
            this.props.pageNumber,
            this.props.jsonSchema,
            this.props.page,
            this.props.fieldNumber
          );
          elements.push(oneComponent.getJsx());
        }
      });
    }
    return elements;
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
    //rebuild children based on drop down selection
    this.reBuildChildren(value);

    this.props.page.loadPageChildren();
  };

  reBuildChildren(value) {
    if (this.props.fieldDef.fields) {
      this.props.fieldDef.fields.forEach(subfield => {
        if (subfield.type === "fieldGroup") {
          if (subfield.displayWhen === value) {
            subfield.visible = true;
          } else {
            subfield.visible = false;
          }
        }
      });
    }
  }

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
