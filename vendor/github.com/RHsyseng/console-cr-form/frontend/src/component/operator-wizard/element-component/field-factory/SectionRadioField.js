import React from "react";

import FieldFactory from "./FieldFactory";

export class SectionRadioField {
  constructor(props) {
    this.props = props;
  }

  /*
   *  TODO: This one is fixed by the sso and LDAP sections, must be dynamic
   */
  getJsx() {
    var section = this.props.fieldDef.label + "section";

    var children = FieldFactory.newInstancesAsJsx(
      this.props.fieldDef.fields,
      this.props.jsonSchema,
      this.props.pageNumber,
      this.props.page,
      this.props.fieldDef.label
    );

    return (
      <div id={section} key={section}>
        {children}
      </div>
    );
  }
}
