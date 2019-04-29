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
    //TODO: we'll figure this out later
    /*
    var j = this.props.pageNumber;
    var cnt = 0;
    if (j < 2) {
      while (j >= 0) {
        cnt = cnt + this.state.jsonForm.pages[j].fields.length + 2;
        j--;
      }
    }
    var elem = document.getElementById("main_form").elements;
    const len = cnt > 0 ? cnt : elem.length;
    var str = "";
    var sampleYaml = {};
    for (var i = 0; i < len; i++) {
      if (elem[i].type != "button") {
        var jsonpath = document
          .getElementById(elem[i].id)
          .getAttribute("jsonpath");
        if (
          elem[i].value != null &&
          elem[i].value != "" &&
          elem[i].name != "alt-form-checkbox-1" &&
          jsonpath != "$.spec.auth.sso" &&
          jsonpath != "$.spec.auth.ldap"
        ) {
          str += "Name: " + elem[i].name + " ";
          str += "Type: " + elem[i].type + " ";
          str += "Value: " + elem[i].value + " ";
          str += "                                                 ";

          var tmpJsonPath = utils.getJsonSchemaPathForYaml(jsonpath);
          const value =
            elem[i].type === "checkbox" ? elem[i].checked : elem[i].value;
            */
    //if (tmpJsonPath.search(/\*/g) != -1) {
    /*
            tmpJsonPath = utils.replaceStarwithPos(elem[i], tmpJsonPath);
          }

          sampleYaml[tmpJsonPath] = value;
        }
      }
    }
    alert(str);
    console.log(sampleYaml);
    var result = this.props.createResultYaml(sampleYaml);
    console.log(result);

    this.props.togglePopup();
    */
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
    //   const jsonObject = {};
    //   const pageDef = this.props.page.props.pageDef;
    //   if (pageDef != null && pageDef != "") {
    //     if (pageDef.fields != null && pageDef.fields != "") {
    //       pageDef.fields.forEach(field => {
    //         if (field.type === "object") {
    //           field.fields.forEach(child => {
    //             console.log(YAML.safeDump(child));
    //           });
    //         }
    //         if (field.value !== undefined && field.value !== "") {
    //           let jsonPath = this.getJsonSchemaPathForYaml(field.jsonPath);

    //           jsonObject[jsonPath] = field.value;
    //         }
    //       });
    //     }

    //     // console.log(YAML.safeDump(jsonObject));
    //     Dot.object(jsonObject);
    //     console.log(YAML.safeDump(Dot.object(jsonObject)));

    //     alert(YAML.safeDump(Dot.object(jsonObject)));
    //   }
    // };

    // getJsonSchemaPathForYaml(jsonPath) {
    //   //console.log("json Path: " + jsonPath);
    //   jsonPath = jsonPath.slice(2, jsonPath.length);

    //   //console.log("jsonSchema Path: " + jsonPath);
    //   return jsonPath;
  };
}
