import React from "react";

import { FormGroup, Radio } from "@patternfly/react-core";

export class RadioButtonField {
  constructor(props) {
    this.props = props;
    this.handleChangeRadio = this.handleChangeRadio.bind(this);
  }

  getJsx() {
    return (
      <FormGroup
        fieldId={this.props.ids.fieldGroupId}
        key={this.props.ids.fieldGroupKey}
        helperText={this.props.fieldDef.description}
      >
        <Radio
          key={this.props.ids.fieldKey}
          defaultValue={this.props.fieldDef.label}
          onChange={this.handleChangeRadio}
          name={this.props.parentid}
          isChecked={this.props.fieldDef.value}
          label={this.props.fieldDef.label}
          id={this.props.fieldDef.label}
        />
      </FormGroup>
    );
  }

  handleChangeRadio = () => {
    this.isCheckedRadio = !this.isCheckedRadio;

    let count = 0,
      pos = 0;
    this.props.page.props.pageDef.fields.forEach((field, i) => {
      if (field.label === this.props.parentid) {
        //locate parent pos
        pos = i;
        field.fields.forEach(subfield => {
          if (subfield.label !== this.props.fieldDef.label) {
            count = subfield.fields.length; //previously added
            subfield.value = false;
          } else {
            subfield.value = true;
          }
        });
      } // );
    });

    //remove
    if (this.props.page.props.pageDef.fields[pos].value !== undefined) {
      this.props.page.props.pageDef.fields.splice(pos + 1, count);
    }

    //add
    if (this.props.fieldDef.fields) {
      this.props.fieldDef.fields.forEach((field, i) => {
        this.props.page.props.pageDef.fields.splice(pos + 1 + i, 0, field);
      });
    }
    this.props.page.props.pageDef.fields[pos].value = this.props.fieldDef.label;

    this.props.page.loadPageChildren();
  };
}
