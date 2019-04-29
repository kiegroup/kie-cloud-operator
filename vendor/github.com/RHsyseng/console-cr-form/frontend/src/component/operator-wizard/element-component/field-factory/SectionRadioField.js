import React from "react";
import { Radio } from "@patternfly/react-core";
import FieldFactory from "./FieldFactory";

export class SectionRadioField {
  constructor(props) {
    this.props = props;
    this.handleChangeRadio = this.handleChangeRadio.bind(this);
    this.state = {
      ssoORldap: ""
    };
  }

  /*
   *  TODO: This one is fixed by the sso and LDAP sections, must be dynamic
   */
  getJsx() {
    var section = this.props.fieldDef.label + "section";

    var isCheckedRadio = this.state.ssoORldap === section;
    var children = FieldFactory.newInstancesAsJsx(
      this.props.fieldDef.fields,
      null
    );

    return (
      <div key={"section-" + this.props.ids.fieldKey}>
        <Radio
          key={this.props.ids.fieldKey}
          value={section}
          isChecked={isCheckedRadio}
          name="ssoOrldap"
          onChange={this.handleChangeRadio}
          label={this.props.fieldDef.label}
          id={this.props.fieldDef.label}
        />
        <div id={section} key={section} style={{ display: "none" }}>
          {children}
        </div>
      </div>
    );
  }

  handleChangeRadio = (checked, event) => {
    const value = event.currentTarget.value;
    this.state = { ssoORldap: value };

    if (value == "LDAPsection") {
      document.getElementById("SSOsection").style.display = "none";
      document.getElementById("LDAPsection").style.display = "block";
    } else {
      document.getElementById("LDAPsection").style.display = "none";
      document.getElementById("SSOsection").style.display = "block";
    }
  };
}
