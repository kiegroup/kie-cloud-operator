import React, { Component } from "react";
//import ReactDOM from "react-dom";
//import { Form } from "@patternfly/react-core";
import {
  FormGroup,
  TextInput,
  TextArea,
  FormSelectOption,
  FormSelect,
  Radio,
  Button,
  ActionGroup,
  Checkbox
} from "@patternfly/react-core";
import { PlusCircleIcon, MinusCircleIcon } from "@patternfly/react-icons";
import validator from "validator";
import JSONPATH from "jsonpath";
//import { OPERATOR_NAME } from "./common/GuiConstants";
import * as utils from "./common/CommonUtils";

export default class PageBase extends Component {
  componentDidMount() {
    this.renderComponents();
  }

  onSubmit = () => {
    console.log("onSubmit is clicked" + this.state.pageNumber);
    console.log(
      "onSubmit is clicked" + this.state.jsonForm.pages[0].fields.length
    );
    var j = this.state.pageNumber;
    var cnt = 0;
    if (j < 2) {
      while (j >= 0) {
        console.log(
          "onSubmit is clicked" + this.state.jsonForm.pages[j].fields.length
        );
        cnt = cnt + this.state.jsonForm.pages[j].fields.length + 2;
        j--;
      }
    }
    //console.log(The no of field needed for yaml creation on editcnt + "cnt");
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
          if (tmpJsonPath.search(/\*/g) != -1) {
            tmpJsonPath = utils.replaceStarwithPos(elem[i], tmpJsonPath);
          }
          //
          sampleYaml[tmpJsonPath] = value;
          //  }
        }
      }
    }
    alert(str);
    console.log(sampleYaml);
    var result = this.props.createResultYaml(sampleYaml);
    console.log(result);
    //  alert(result);
    //  this.props.setResultYaml(result);
    this.props.togglePopup();
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

  onChange = value => {
    console.log("onChange with value: " + value);
  };

  onChangeCheckBox = (_, event) => {
    const target = event.target;
    const value = target.type === "checkbox" ? target.checked : target.value;

    this.setState({ [event.target.name]: value });
  };

  //onChangeEmail = (value, event) => {
  onChangeEmail = value => {
    //console.log("onChangeEmail: " + value);
    //console.log("handleEmailchange, event.target.name: " + event.target.name);
    //console.log("handleEmailchange, event.target.value: " +  event.target.value);

    if (value != null && value != "" && !validator.isEmail(value)) {
      console.log("not valid email address: " + value);
      this.setState({
        validationMessageEmail: "not valid email address"
      });
    } else {
      this.setState({
        validationMessageEmail: ""
      });
    }
  };

  //onChangeUrl = (value, event) => {
  onChangeUrl = value => {
    if (value != null && value != "" && !validator.isURL(value)) {
      console.log("not valid URL " + value);
      this.setState({
        validationMessageUrl: "not valid URL"
      });
    } else {
      this.setState({
        validationMessageUrl: ""
      });
    }
  };

  findValueFromSchema(jsonPath) {
    const schema = this.props.jsonSchema;
    //const values = schema.properties.spec.properties.environment.enum;
    //console.log("values " + JSON.stringify(values));

    //console.log("jsonPath: " + jsonPath);
    //passed in: $.spec.environment
    //jsonPath = "$..spec.properties.environment.enum";

    var queryResults = JSONPATH.query(schema, jsonPath);
    //console.log("queryResults " + JSON.stringify(queryResults[0]));

    return queryResults[0];
  }

  renderComponents = () => {
    //const pageDef = MockupData_JSON.pages[1];
    //console.log("!!!!!!!!!! renderComponents pageDef2: " + JSON.stringify(pageDef));
    const pageDef = this.state.jsonForm.pages[this.state.pageNumber];

    if (pageDef != null && pageDef != "") {
      var children = [];

      const tmpDiv = (
        <b key={this.state.pageNumber}>PAGE {this.state.pageNumber + 1}</b>
      );
      children.push(tmpDiv);
      //generate all fields
      if (pageDef.fields != null && pageDef.fields != "") {
        //loop through all fields
        pageDef.fields.forEach((field, fieldNumber) => {
          const oneComponent = this.buildOneField(field, fieldNumber);
          children.push(oneComponent);
        });
      }

      //generate all buttons
      if (pageDef.buttons != null && pageDef.buttons != "") {
        const buttonsComponent = this.buildAllButtons(pageDef.buttons);
        children.push(buttonsComponent);
      }

      //return children;
      this.setState({ children });
    } else {
      console.log("do nothing, it's an empty page.");
      //do nothing, it's an empty page.
    }
  };

  retrieveObjectMap(fieldNumber) {
    const key = this.state.pageNumber + "_" + fieldNumber;
    // const key=fieldGroupKey;
    var value = this.state.objectMap.get(key);

    if (value == null) {
      return "";
    } else {
      return JSON.parse(value);
    }
  }

  storeObjectMap(field, fieldNumber) {
    //first time deal with this key value pair, store fields (the whole array, can't be just field[0]) to the map
    const key = this.state.pageNumber + "_" + fieldNumber;
    // const key=fieldGroupKey;

    this.state.objectMap.set(key, JSON.stringify(field.fields));
    this.state.objectCntMap.set(key, field.fields.length);
  }

  deleteOneFieldForObj = event => {
    var fieldNumber = document
      .getElementById(event.target.id)
      .getAttribute("fieldnumber");
    //console.log(      "deleteOneFieldForObj, fieldNumber : " + JSON.parse(fieldNumber)    );

    var parentFieldNumber = document
      .getElementById(event.target.id)
      .getAttribute("parentfieldnumber");
    //console.log(      "deleteOneFieldForObj, parentFieldNumber : " +        JSON.parse(parentFieldNumber)    );

    var field;
    if (parentFieldNumber == -1) {
      field = this.state.jsonForm.pages[this.state.pageNumber].fields[
        fieldNumber
      ];
    } else {
      field = this.state.jsonForm.pages[this.state.pageNumber].fields[
        parentFieldNumber
      ].fields[fieldNumber];
    }
    //console.log("deleteOneFieldForObj, field.min current value: " + field.min);

    var combinedFieldNumber = parentFieldNumber + "_" + fieldNumber;
    const sampleObj = this.retrieveObjectMap(combinedFieldNumber);
    //console.log("deleteOneFieldForObj, sampleObj length: " + sampleObj.length);

    if (field.min > 0) {
      for (var i = 0; i < sampleObj.length; i++) {
        field.fields.pop();
      }

      field.min = field.min - 1;
      this.renderComponents();
    } else {
      console.log("deleteOneFieldForObj, min = 0, can't delete more!");
    }
  };

  addOneFieldForObj = event => {
    var fieldNumber = document
      .getElementById(event.target.id)
      .getAttribute("fieldnumber");
    //console.log("addOneFieldForObj, fieldNumber: " + JSON.parse(fieldNumber));

    var parentFieldNumber = document
      .getElementById(event.target.id)
      .getAttribute("parentfieldnumber");
    //console.log(      "addOneFieldForObj, parentFieldNumber: " + JSON.parse(parentFieldNumber)    );

    var field;
    //parentFieldNumber == -1 means it's 1st tier object field
    if (parentFieldNumber == -1) {
      field = this.state.jsonForm.pages[this.state.pageNumber].fields[
        fieldNumber
      ];
    } else {
      field = this.state.jsonForm.pages[this.state.pageNumber].fields[
        parentFieldNumber
      ].fields[fieldNumber];
    }

    //console.log("addOneFieldForObj, field.min current value: " + field.min);

    var combinedFieldNumber = parentFieldNumber + "_" + fieldNumber;
    const sampleObj = this.retrieveObjectMap(combinedFieldNumber);

    if (field.min < field.max) {
      //console.log("addOneFieldForObj, min < max, add another object");
      field.min = field.min + 1;
      //console.log("addOneFieldForObj, field.min new value:" + field.min);

      //the whole idea about this seperateObjDiv is to make the screen looks cleaner when user add a new obj
      //const seperateObjDiv = JSON.parse('{"type":"seperateObjDiv"}');
      field.fields = field.fields.concat(sampleObj);
      //field.fields = field.fields.concat(seperateObjDiv);

      this.props.saveJsonForm(this.state.jsonForm);
    } else {
      console.log("addOneFieldForObj, min = max, can't add more!");
    }
    this.renderComponents();
  };

  buildObject(field, fieldNumber, parentFieldNumber) {
    /*
    console.log(" field " + field.label);

    console.log(" fieldNumber " + fieldNumber);
    console.log(" parentFieldNumber " + parentFieldNumber);
*/
    var randomNum = Math.floor(Math.random() * 100000000 + 1);

    const fieldGroupId =
      this.state.pageNumber +
      "_fieldGroup_" +
      fieldNumber +
      "_" +
      parentFieldNumber +
      "_" +
      field.label +
      "_";
    randomNum;
    const fieldGroupKey = "fieldGroupKey_" + fieldGroupId;
    const fieldId =
      this.state.pageNumber +
      "_field_" +
      fieldNumber +
      "_" +
      parentFieldNumber +
      "_" +
      field.label +
      "_" +
      randomNum;
    const fieldKey = "fieldKey_" + fieldId;
    var jsxArray = [];
    var fieldJsx;

    //parentFieldNumber == -1   means this the 1st tier obj field, no parent field.
    if (parentFieldNumber == -1) {
      jsxArray.push(
        <div key={fieldGroupId + 1}>
          <br />
          <br />
          <br />
          <br />
        </div>
      );
    }

    fieldJsx = (
      <ActionGroup fieldid={fieldGroupId} key={fieldGroupKey}>
        <Button
          variant="secondary"
          id={fieldId}
          key={fieldKey}
          fieldnumber={fieldNumber}
          parentfieldnumber={parentFieldNumber}
          onClick={this.addOneFieldForObj}
        >
          Add new {field.label}
        </Button>
        <Button
          variant="secondary"
          id={fieldId + 1}
          key={fieldKey + 1}
          fieldnumber={fieldNumber}
          parentfieldnumber={parentFieldNumber}
          onClick={this.deleteOneFieldForObj}
        >
          Delete last {field.label}
        </Button>
      </ActionGroup>
    );

    jsxArray.push(fieldJsx);
    /*
    jsxArray.push(
      <div key={fieldGroupId + 1}>
        <br />
        <br />
        <br />
        <br />
      </div>
  );*/
    var combinedFieldNumber = parentFieldNumber + "_" + fieldNumber;
    //console.log("combinedFieldNumber " + combinedFieldNumber);

    var sampleObj = this.retrieveObjectMap(combinedFieldNumber);
    const objCnt =
      this.retrieveObjectCntMap(field, fieldNumber, fieldGroupKey) + 1; //getting how many field in obj e.g env has 2 name and value +1 for devider
    if (sampleObj == "") {
      //it's the first time here, no such sample in the objectMap yet, so store it.
      this.storeObjectMap(field, combinedFieldNumber);

      if (parentFieldNumber != -1) {
        //this is a 2nd tier object, double check if its parent is in the map or not.
        const parentCombinedFieldNumber = "-1" + "_" + parentFieldNumber;
        const parentSampleObj = this.retrieveObjectMap(
          parentCombinedFieldNumber
        );
        if (parentSampleObj == "") {
          //its parent is not store in object map yet, store the sample now.
          const parentField = this.state.jsonForm.pages[this.state.pageNumber]
            .fields[parentFieldNumber];
          this.storeObjectMap(parentField, parentCombinedFieldNumber);

          if (parentField.min == 0) {
            parentField.fields = [];
          }
        }
      }

      if (field.min == 0) {
        //if it's the 1st time here, and field.min ==0
        //so after store it to the map, remove from render json form, then it won't be displayed
        field.fields = [];
      } else if (field.min > 1) {
        //for field.min == 1 do nothing, just leave the sample there as the 1st object in array which will be displayed
        //for field.min > 1, need to insert more objects as the min value requires
        //TODO:

        console.log("!!!!!!!! TODO: add more objects");
      }
    }

    var pos = 0,
      cnt = 1,
      attrs = {};
    field.fields.forEach((subfield, i) => {
      if (field.min == 0) {
        //means don't generate the 1st one unless user press button
        //console.log("field.min == 0, won't render ");
      } else {
        //add extra attributes in fields to recognise the position of field are created  envpos, serverpos etc.
        // these attr will be replace *  by serverpos and envpos in jsonpath like  $.spec.objects.servers[*].env[*].value
        var posKey = field.label.toLowerCase() + "pos";
        attrs = {
          ...attrs,
          [posKey]: pos
        };

        //console.log("!!!!!!!! here3, subfield: " + JSON.stringify(subfield));
        if (subfield.type != "object") {
          let oneComponent = this.buildOneField(subfield, i);
          jsxArray.push(oneComponent);
        } else {
          let oneComponent = this.buildObject(subfield, i, fieldNumber);
          jsxArray.push(oneComponent);
        }

        //assigning each fiels for object with same pos and increment the pos when all fields are done
        if (cnt == objCnt) {
          //console.log("Incrementing pos from " + pos + " to " + (pos + 1));
          pos++;
          cnt = 0;
        }
        cnt++;
      }
    });
    if (parentFieldNumber == -1) {
      jsxArray.push(
        <div key={fieldGroupId + 2}>
          <br />
          <br />
          <br />
          <br />
        </div>
      );
    }

    return jsxArray;
  }

  retrieveObjectCntMap(field, fieldNumber) {
    const key = this.state.pageNumber + "_" + fieldNumber;
    var value = this.state.objectCntMap.get(key);

    // console.log(
    //   "retrieveObjectCntMap value::::::::: " + key + " : " + value
    // );

    if (value == null) {
      return "";
    } else {
      return value;
    }
  }
  buildOneField(field, fieldNumber, attrs) {
    var randomNum = Math.floor(Math.random() * 100000000 + 1);

    const fieldGroupId =
      this.state.pageNumber +
      "_fieldGroup_" +
      fieldNumber +
      "_" +
      field.label +
      "_" +
      randomNum;
    const fieldGroupKey = "fieldGroupKey_" + fieldGroupId;
    const fieldId =
      this.state.pageNumber +
      "_field_" +
      fieldNumber +
      "_" +
      field.label +
      "_" +
      randomNum;
    const fieldKey = "fieldKey_" + fieldId;
    const textName = field.label;

    var fieldJsx = "";
    if (field.type == "dropDown") {
      var options = [];
      const tmpJsonPath = utils.getJsonSchemaPathForJsonPath(field.jsonPath);
      const optionValues = this.findValueFromSchema(tmpJsonPath + ".enum");
      if (optionValues !== undefined) {
        optionValues.forEach(option => {
          const oneOption = {
            label: option,
            value: option
          };
          options.push(oneOption);
        });
      }
      //  const tmpJsonPath = "$..spec.properties.environment.description";
      const helpText = this.findValueFromSchema(tmpJsonPath + ".description");

      fieldJsx = (
        <FormGroup
          label={field.label}
          fieldId={fieldGroupId}
          key={fieldGroupKey}
          helperText={helpText}
        >
          <FormSelect
            id={fieldId}
            key={fieldKey}
            name={textName}
            jsonpath={field.jsonPath}
          >
            {options.map((option, index) => (
              <FormSelectOption
                isDisabled={option.disabled}
                id={fieldId + index}
                key={fieldKey + index}
                value={option.value}
                label={option.label}
              />
            ))}
          </FormSelect>
        </FormGroup>
      );
    } else if (field.type == "textArea") {
      fieldJsx = (
        <FormGroup
          label={field.label}
          isRequired
          fieldId={fieldGroupId}
          key={fieldGroupKey}
        >
          <TextArea
            value={field.default}
            name="horizontal-form-exp"
            id={fieldId}
            key={fieldKey}
            jsonpath={field.jsonPath}
          />
        </FormGroup>
      );
    } else if (field.type == "radioButton") {
      const fieldIdTrue = fieldId + "-true";
      const fieldKeyTrue = fieldKey + "-true";
      const fieldIdFalse = fieldId + "-false";
      const fieldKeyFalse = fieldKey + "-false";
      fieldJsx = (
        <FormGroup
          label={field.label}
          fieldId={fieldGroupId}
          key={fieldGroupKey}
        >
          <Radio
            label="Yes"
            aria-label="radio yes"
            id={fieldIdTrue}
            key={fieldKeyTrue}
            name="horizontal-radios"
            jsonpath={field.jsonPath}
          />
          <Radio
            label="No"
            aria-label="radio no"
            id={fieldIdFalse}
            key={fieldKeyFalse}
            name="horizontal-radios"
            jsonpath={field.jsonPath}
          />
        </FormGroup>
      );
    } else if (field.type == "object") {
      fieldJsx = this.buildObject(field, fieldNumber, -1);
    } else if (field.type == "email") {
      fieldJsx = (
        <FormGroup
          label={field.label}
          fieldId={fieldGroupId}
          key={fieldGroupKey}
        >
          <TextInput
            type="text"
            id={fieldId}
            key={fieldKey}
            name={textName}
            onChange={this.onChangeEmail}
            jsonpath={field.jsonPath}
          />
        </FormGroup>
      );
    } else if (field.type == "url") {
      fieldJsx = (
        <FormGroup
          label={field.label}
          fieldId={fieldGroupId}
          key={fieldGroupKey}
        >
          <TextInput
            type="text"
            id={fieldId}
            key={fieldKey}
            name={textName}
            onChange={this.onChangeUrl}
            jsonpath={field.jsonPath}
          />
        </FormGroup>
      );
    } else if (field.type == "password") {
      fieldJsx = (
        <FormGroup
          label={field.label}
          fieldId={fieldGroupId}
          key={fieldGroupKey}
        >
          <TextInput
            type="password"
            id={fieldId}
            key={fieldKey}
            name={textName}
            onChange={this.onChange}
            jsonpath={field.jsonPath}
          />
        </FormGroup>
      );
    } else if (field.type == "checkbox") {
      var name = "checkbox-" + fieldNumber;
      var isChecked = field.default == "true" || field.default == "TRUE";
      fieldJsx = (
        <FormGroup
          label={field.label}
          fieldId={fieldGroupId}
          key={fieldGroupKey}
        >
          <Checkbox
            isChecked={isChecked}
            onChange={this.onChangeCheckBox}
            id={fieldId}
            key={fieldKey}
            aria-label="checkbox yes"
            name={name}
            jsonpath={field.jsonPath}
          />
        </FormGroup>
      );
    } else if (field.type == "seperateObjDiv") {
      fieldJsx = (
        <div key={fieldKey}>
          <hr />
        </div>
      );
    } else if (field.type == "section") {
      fieldJsx = this.buildSection(fieldNumber);
    } else if (field.type == "section_radio") {
      fieldJsx = this.buildMutualExclusiveObject(fieldNumber);
    } else {
      fieldJsx = (
        <FormGroup
          label={field.label}
          fieldId={fieldGroupId}
          key={fieldGroupKey}
        >
          <TextInput
            isRequired
            type="text"
            id={fieldId}
            key={fieldKey}
            aria-describedby="horizontal-form-name-helper"
            name={textName}
            onChange={this.onChange}
            jsonpath={field.jsonPath}
            {...attrs}
          />
        </FormGroup>
      );
    }

    return fieldJsx;
  }

  buildSection(fieldNumber) {
    const field = this.state.jsonForm.pages[this.state.pageNumber].fields[
      fieldNumber
    ];

    var randomNum = Math.floor(Math.random() * 100000000 + 1);
    const fieldGroupId =
      this.state.pageNumber +
      "_fieldGroup_" +
      fieldNumber +
      "_" +
      field.label +
      "_" +
      randomNum;
    const fieldGroupKey = "fieldGroupKey_" + fieldGroupId;
    const fieldId =
      this.state.pageNumber +
      "_field_" +
      fieldNumber +
      "_" +
      field.label +
      "_" +
      randomNum;
    const fieldKey = "fieldKey_" + fieldId;
    var section = field.label + "section";
    var jsxArray = [];
    var iconIdPlus = section + "plus";
    var iconIdMinus = section + "minus";

    var fieldJsx = (
      <FormGroup fieldId={fieldGroupId} key={fieldGroupKey}>
        <Button
          variant="link"
          id={fieldId}
          key={fieldKey}
          fieldNumber={fieldNumber}
          onClick={this.expandSection}
          name={section}
          style={{ display: "inline-block" }}
        >
          {field.label}
          <PlusCircleIcon id={iconIdPlus} style={{ display: "block" }} />{" "}
          <MinusCircleIcon id={iconIdMinus} style={{ display: "none" }} />
        </Button>
      </FormGroup>
    );

    //console.log("fieldNumber Section" + fieldNumber);

    var children = [];
    if (field.fields != null && field.fields != "") {
      //loop through all fields
      field.fields.forEach((subfield, i) => {
        let oneComponent = this.buildOneField(subfield, i);
        children.push(oneComponent);
      });
    }
    jsxArray.push(fieldJsx);
    fieldJsx = (
      <div id={section} key={section} style={{ display: "none" }}>
        {children}
      </div>
    );

    jsxArray.push(fieldJsx);
    console.log(jsxArray);
    return jsxArray;
  }

  buildMutualExclusiveObject(fieldNumber) {
    const field = this.state.jsonForm.pages[this.state.pageNumber].fields[
      fieldNumber
    ];

    // var randomNum = Math.floor(Math.random() * 100000000 + 1);

    // const fieldId =
    //   this.state.pageNumber +
    //   "_field_" +
    //   fieldNumber +
    //   "_" +
    //   field.label +
    //   "_" +
    //   randomNum;
    //const fieldKey = "fieldKey_" + fieldId;
    var section = field.label + "section";
    var jsxArray = [];

    // var name = "checkbox-" + fieldNumber;
    // var isChecked = field.default == "true" || field.default == "TRUE";
    // this.setState({ ssoORldap:section });
    var isCheckedRadio = this.state.ssoORldap === section;
    // console.log(
    //   "isCheckedRadioisCheckedRadio" + isCheckedRadio + this.state.ssoORldap
    // );
    var fieldJsx = (
      <Radio
        value={section}
        isChecked={isCheckedRadio}
        name="ssoOrldap"
        onChange={this.handleChangeRadio}
        label={field.label}
        id={field.label}
      />
    );
    var children = [];
    if (field.fields != null && field.fields != "") {
      //loop through all fields
      field.fields.forEach((subfield, i) => {
        let oneComponent = this.buildOneField(subfield, i);
        children.push(oneComponent);
      });
    }
    jsxArray.push(fieldJsx);
    fieldJsx = (
      <div id={section} key={section} style={{ display: "none" }}>
        {children}
      </div>
    );

    jsxArray.push(fieldJsx);
    console.log(jsxArray);
    return jsxArray;
  }

  handleChangeRadio = (checked, event) => {
    /// const target = event.target;
    const value = event.currentTarget.value;
    this.setState({ ssoORldap: value });
    // const name = target.name;

    //var elem = document.getElementById(value);
    if (value == "LDAPsection") {
      document.getElementById("SSOsection").style.display = "none";
      document.getElementById("LDAPsection").style.display = "block";
    } else {
      document.getElementById("LDAPsection").style.display = "none";
      document.getElementById("SSOsection").style.display = "block";
    }
    // if (value.search(event.target.id)!=-1){
    //         elem.style.display="block";

    //        }
    //         else{
    //         elem.style.display="none";

    //         }
  };

  expandSection = event => {
    const target = event.target;
    const name = target.name;

    //this.setState({ [name]: value });
    var elem = document.getElementById(name);
    console.log(elem);

    if (elem.style.display === "block") {
      elem.style.display = "none";
      document.getElementById(name + "plus").style.display = "block";
      document.getElementById(name + "minus").style.display = "none";
    } else {
      elem.style.display = "block";
      document.getElementById(name + "plus").style.display = "none";
      document.getElementById(name + "minus").style.display = "block";
    }
  };

  buildAllButtons(buttons) {
    var buttonsJsx = [];
    //loop through all
    buttons.forEach((button, i) => {
      const key = this.state.pageNumber + "-form-key-" + button.label + i;

      var buttonJsx = "";
      if (button.action != null && button.action == "submit") {
        buttonJsx = (
          <Button variant="primary" key={key} onClick={this.onSubmit}>
            {button.label}
          </Button>
        );
      } else if (button.action != null && button.action == "cancel") {
        buttonJsx = (
          <Button variant="secondary" key={key} onClick={this.onCancel}>
            {button.label}
            Component
          </Button>
        );
      } else if (button.action != null && button.action == "next") {
        buttonJsx = (
          <Button variant="secondary" key={key} onClick={this.onNext}>
            {button.label}
          </Button>
        );
      } else if (button.action != null && button.action == "close") {
        buttonJsx = (
          <Button variant="secondary" key={key} onClick={this.onClose}>
            {button.label}
          </Button>
        );
      }

      buttonsJsx.push(buttonJsx);
    });

    const actionGroupKey = this.state.pageNumber + "-action-group";
    return <ActionGroup key={actionGroupKey}>{buttonsJsx}</ActionGroup>;
  }

  render() {
    return <div>{this.state.children}</div>;
  }

  /*
  renderMultipleComponents = (label, operator, tempRenderJson) => {
    const objDef = objJson[operator + "_" + label];

    console.log(
      "!!!!!!!!!! renderComponents tempRenderJson objDefobjDef***: " +
        JSON.stringify(objDef)
    );
    //
    // var obj = JSON.parse(jsonStr);
    var x;
    let obj = objDef.fields;

    for (x in obj) {
      tempRenderJson.push(obj[x]);
      // jsonStr = JSON.stringify(obj);
    }
    // this.state.children = [];
    return tempRenderJson;
  };

  onAddObject = () => {
    var x, currentFields, fields;
    currentFields = this.state.renderJson.fields;
    console.log("onAddObject is clicked****", currentFields);
    var tempRenderJson = [];

    for (x in currentFields) {
      fields = currentFields[x];
      tempRenderJson.push(fields);

      // if (fields.label === "Env") {
      if (fields.type == "object") {
        //   console.log("onAddObject is clicked****", fields.type);
        this.renderMultipleComponents(
          fields.label,
          OPERATOR_NAME,
          tempRenderJson
        );
      }
    }
    // console.log(
    //   "!!!!!!!!!! renderComponents tempRenderJson:::: " +
    //     JSON.stringify(tempRenderJson)
    // );
    let newRenderJson = {
      fields: [{}],
      buttons: [{}]
    };
    // console.log(
    //   "!!!!!!!!!! renderComponents newRenderJs+===== " +
    //     JSON.stringify(newRenderJson.fields)
    // );
    newRenderJson = { ...newRenderJson, fields: tempRenderJson };
    newRenderJson = {
      ...newRenderJson,
      buttons: this.state.pageDef.buttons
    };

    // console.log(
    //   "!!!!!!!!!! renderComponents newRenderJson:::: " +
    //     JSON.stringify(newRenderJson)
    // );

    this.setState({ renderJson: newRenderJson });
    //  this.state.children = [];

    this.renderComponents(newRenderJson);
  };

  onDeleteObject = () => {
    // console.log(
    //   "!!!!!!!!!! renderComponents newRenderJson:::: " +
    //     JSON.stringify(this.state.renderJson)
    // );

    var tempRenderJson = this.state.renderJson;
    var newChildren = [];
    tempRenderJson.fields.splice(tempRenderJson.fields.length - 2, 2);
    this.setState({ renderJson: tempRenderJson });
    //  this.state.children = [];
    this.setState({
      children: [this.state.children, newChildren]
    });

    this.renderComponents(tempRenderJson);
};*/
}
