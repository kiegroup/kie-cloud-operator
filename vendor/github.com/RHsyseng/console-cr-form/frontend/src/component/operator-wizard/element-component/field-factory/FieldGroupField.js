import React from "react";

import { FormGroup } from "@patternfly/react-core";

import FieldFactory from "./FieldFactory";
import {
  ENV_KEY,
  GITHOOKS_FIELD,
  CONSOLE_STEP,
  GITHOOKS_ENVS
} from "../../../common/GuiConstants";

export class FieldGroupField {
  constructor(props) {
    this.props = props;
    this.children = [];
    this.addChildren = this.addChildren.bind(this);
    this.parentFieldNumber =
      this.props.parentid === undefined ? -1 : this.props.parentid;
    this.grandParentFieldNumber = this.props.grandParentId
      ? this.props.grandParentId
      : -1;
  }

  getJsx() {
    var section = this.props.fieldDef.label + "section";
    var jsxArray = [];
    if (
      this.props.page !== undefined &&
      this.props.page.props.getObjectMap(ENV_KEY) !== undefined &&
      this.props.page.props.pageDef.label === CONSOLE_STEP &&
      this.props.fieldDef.label === GITHOOKS_FIELD
    ) {
      if (
        GITHOOKS_ENVS.indexOf(this.props.page.props.getObjectMap(ENV_KEY)) > -1
      ) {
        this.props.fieldDef.visible = true;
      } else {
        this.props.fieldDef.visible = false;
      }
    }

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
        <div style={{ fontWeight: "bold" }}>{this.props.fieldDef.label}</div>

        <div className="pf-c-card">
          <div className="pf-c-card__body pf-c-form">{this.children}</div>
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
        if (subfield.type != "object" && subfield.type != "fieldGroup") {
          let oneComponent = FieldFactory.newInstance(
            subfield,
            i,
            this.props.pageNumber,
            this.props.jsonSchema,

            this.props.page
          );
          elements.push(oneComponent);
        } else {
          console.log(
            "parentId" +
              this.props.fieldNumber +
              " grandParentId" +
              this.props.parentid
          );
          let oneComponent = FieldFactory.newInstance(
            subfield,
            i,
            this.props.pageNumber,
            this.props.jsonSchema,
            this.props.page,
            this.props.fieldNumber,
            this.props.parentid
          );
          elements.push(oneComponent);
        }
      });
    }
    return elements;
  }
}
