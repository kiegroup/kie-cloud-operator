import React from "react";
import validator from "validator";
import { FormGroup, TextInput } from "@patternfly/react-core";

export class EmailField {
  constructor(props) {
    this.props = props;
  }

  getJsx() {
    return (
      <FormGroup
        label={this.props.fieldDef.label}
        fieldId={this.props.ids.fieldGroupId}
        key={this.props.ids.fieldGroupKey}
      >
        <TextInput
          type="text"
          id={this.props.ids.fieldId}
          key={this.props.ids.fieldKey}
          name={this.props.fieldDef.label}
          onChange={this.onChangeEmail}
          jsonpath={this.props.fieldDef.jsonPath}
        />
      </FormGroup>
    );
  }

  onChangeEmail = value => {
    if (value != null && value != "" && !validator.isEmail(value)) {
      console.log("not valid email address: " + value);
      /*TODO: fix
      this.setParentState({
        validationMessageEmail: "not valid email address"
      });*/
    } else {
      /*
      this.setParentState({
        validationMessageEmail: ""
      });*/
    }
  };
}
