import React from "react";

import { FormGroup, Checkbox } from "@patternfly/react-core";

export class CheckboxField {
  constructor(props) {
    this.props = props;
  }

  getJsx() {
    var name = "checkbox-" + this.props.fieldNumber;

    return (
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
  }

  onChangeCheckBox = (_, event) => {
    const target = event.target;
    const value = target.type === "checkbox" ? target.checked : target.value;
    this.props.fieldDef.checked = value;
    //  this.setParentState({ [event.target.name]: value });
  };
}
