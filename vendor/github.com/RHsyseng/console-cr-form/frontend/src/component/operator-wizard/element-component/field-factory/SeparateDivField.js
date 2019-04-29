import React from "react";

export class SeparateDivField {
  constructor(props) {
    this.props = props;
  }

  getJsx() {
    return (
      <div key={this.props.ids.fieldKey}>
        ------------------------------------------------------------------------------------------------------------------
      </div>
    );
  }
}
