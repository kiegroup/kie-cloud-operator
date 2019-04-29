import React from "react";
import { FormGroup, TextArea } from "@patternfly/react-core";

export class TextAreaField {
  constructor(props) {
    this.props = props;
  }

  getJsx() {
    return (
      <FormGroup
        label={this.props.fieldDef.label}
        isRequired
        fieldId={this.props.ids.fieldGroupId}
        key={this.props.ids.fieldGroupKey}
      >
        <TextArea
          value={this.props.fieldDef.default}
          name="horizontal-form-exp"
          id={this.props.ids.fieldId}
          key={this.props.ids.fieldKey}
          jsonpath={this.props.fieldDef.jsonPath}
        />
      </FormGroup>
    );
  }
}
