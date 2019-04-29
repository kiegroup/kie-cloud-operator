import React from "react";
import validator from "validator";
import { FormGroup, TextInput } from "@patternfly/react-core";

export class UrlField {
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
          onChange={this.onChangeUrl}
          jsonpath={this.props.fieldDef.jsonPath}
        />
      </FormGroup>
    );
  }

  onChangeUrl = value => {
    if (value != null && value != "" && !validator.isURL(value)) {
      console.log("not valid URL " + value);
      this.setParentState({
        validationMessageUrl: "not valid URL"
      });
    } else {
      this.setParentState({
        validationMessageUrl: ""
      });
    }
  };
}
