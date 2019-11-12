import React, { Component } from "react";

import { FormGroup, Checkbox } from "@patternfly/react-core";
import FieldFactory from "./FieldFactory";
export class CheckboxField extends Component {
  constructor(props) {
    super(props);
    this.props = props;
  }

  getJsx() {
    var name = "checkbox-" + this.props.fieldNumber;
    var jsxArray = [];
    jsxArray.push(
      <FormGroup
        label={this.props.fieldDef.label}
        fieldId={this.props.ids.fieldGroupId}
        key={this.props.ids.fieldGroupKey}
        helperText={this.props.fieldDef.description}
      >
        <Checkbox
          defaultChecked={this.props.fieldDef.checked}
          onChange={this.onChangeCheckBox}
          id={this.props.ids.fieldId}
          key={this.props.ids.fieldKey}
          aria-label="checkbox yes"
          name={name}
          jsonpath={this.props.fieldDef.jsonPath}
        />
      </FormGroup>
    );
    jsxArray.push(this.addChildren());
    return jsxArray;
  }
  reBuildChildren(value) {
    if (this.props.fieldDef.fields) {
      this.props.fieldDef.fields.forEach(subfield => {
        if (subfield.type === "fieldGroup") {
          if (subfield.displayWhen === value.toString()) {
            subfield.visible = true;
          } else {
            subfield.visible = false;
          }
        }
      });
    }
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
          if (parentjsonpath.length <= subfield.jsonPath.length) {
            res = subfield.jsonPath.substring(
              parentjsonpath.length,
              subfield.jsonPath.length
            );
            subfield.jsonPath = parentjsonpath.concat(res);
          }
        }
        if (subfield.type != "object" && subfield.type != "fieldGroup") {
          let oneComponent = FieldFactory.newInstance(
            subfield,
            i,
            this.props.pageNumber,
            this.props.jsonSchema,
            this.props.props.page
          );
          elements.push(oneComponent);
        } else {
          if (subfield.label === this.props.fieldDef.value) {
            subfield.visible = true;
          }
          let oneComponent = FieldFactory.newInstance(
            subfield,
            i,
            this.props.pageNumber,
            this.props.jsonSchema,
            this.props.props.page,
            this.props.fieldNumber
          );
          elements.push(oneComponent);
        }
      });
    }
    return elements;
  }

  onChangeCheckBox = (_, event) => {
    const target = event.target;
    const value = target.type === "checkbox" ? target.checked : target.value;
    this.props.fieldDef.checked = value;
    this.reBuildChildren(value);
    if (this.props.props.page !== undefined) {
      this.props.props.page.loadPageChildren();
    }
  };

  render() {
    return this.getJsx();
  }
}
