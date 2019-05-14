import React from "react";
import { Button } from "@patternfly/react-core";

const BUTTON_ACTION = {
  submit: "submit",
  cancel: "cancel",
  next: "next",
  close: "close",
  editYaml: "editYaml"
};

export class ButtonElement {
  key = "";

  /**
   * Default constructor for the ButtonElement
   * @param {*} props properties to build the button JSX {pageNumber, buttonDef, buttonId}
   */
  constructor(props) {
    this.props = props;
    this.key =
      props.pageNumber + "-form-key-" + props.buttonDef.label + props.buttonId;
    this.onCancel = this.onCancel.bind(this);
    this.onClose = this.onClose.bind(this);
    this.onNext = this.onNext.bind(this);
    this.onSubmit = this.onSubmit.bind(this);
  }

  getJsx() {
    var clickEvent;
    var buttonRole = "secondary";
    switch (this.props.buttonDef.action) {
      case BUTTON_ACTION.submit:
        buttonRole = "primary";
        clickEvent = this.onSubmit;
        break;
      case BUTTON_ACTION.cancel:
        clickEvent = this.onCancel;
        break;
      case BUTTON_ACTION.next:
        clickEvent = this.onNext;
        break;
      case BUTTON_ACTION.close:
        clickEvent = this.onClose;
        break;
      case BUTTON_ACTION.editYaml:
        clickEvent = this.onEditYaml;
        break;
    }
    return (
      <Button variant={buttonRole} key={this.key} onClick={clickEvent}>
        {this.props.buttonDef.label}
      </Button>
    );
  }

  onSubmit = () => {
    alert("here");
    console.log(this.props);
  };

  onCancel = () => {
    console.log("onCancel is clicked");
    alert("onCancel is clicked");
  };

  onNext = () => {
    console.log("onNext is clicked");
    alert("onNext is clicked");
  };

  onClose = () => {
    console.log("onClose is clicked");
    alert("onClose is clicked");
  };

  onEditYaml = () => {
    this.props.page.editYaml();
  };
}
