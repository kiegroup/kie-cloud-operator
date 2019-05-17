import React from "react";

import { FormGroup } from "@patternfly/react-core";

import FieldFactory from "./FieldFactory";

export class FieldGroupField {
  constructor(props) {
    this.props = props;
    this.children = [];
    this.addChildren = this.addChildren.bind(this);
  }

  getJsx() {
    var section = this.props.fieldDef.label + "section";
    var jsxArray = [];

    var fieldJsx = (
      <FormGroup
        fieldId={this.props.ids.fieldGroupId}
        key={this.props.ids.fieldGroupKey}
      />
    );
    jsxArray.push(fieldJsx);

    this.children = this.addChildren();

    fieldJsx = (
      <div
        id={section}
        key={section}
        style={{
          display:
            this.props.fieldDef.visible !== undefined &&
            this.props.fieldDef.visible !== false
              ? "block"
              : "none"
        }}
      >
        <br />
        <div style={{ fontWeight: "bold" }}>{this.props.fieldDef.label}</div>

        <div className="pf-c-card">
          <div className="pf-c-card__body">{this.children}</div>
        </div>
      </div>
    );
    jsxArray.push(fieldJsx);
    return jsxArray;
  }

  addChildren() {
    var pos = 0,
      elements = [];

    if (this.props.fieldDef.fields) {
      this.props.fieldDef.fields.forEach((subfield, i) => {
        if (
          subfield.parent === undefined ||
          (subfield.parent !== undefined &&
            subfield.parent === this.props.fieldDef.value)
        ) {
          var parentjsonpath = this.props.fieldDef.jsonPath;
          var res = "";

          if (parentjsonpath.length < subfield.jsonPath.length) {
            res = subfield.jsonPath.substring(
              parentjsonpath.length,
              subfield.jsonPath.length
            );
            res = parentjsonpath.concat(res);

            subfield.jsonPath = res.replace(/\*/g, pos);
          }
        } else {
          subfield.jsonPath = subfield.jsonPath.replace(/\*/g, pos);
        }
        subfield.visible = this.props.fieldDef.visible;
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
          // console.log("parentId" + this.props.fieldNumber);
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
}
