import React, { Component } from "react";
import ElementFactory from "../element-component/ElementFactory";
import { Form } from "@patternfly/react-core";
import YAML from "js-yaml";
import Dot from "dot-object";
import { TextArea, Button, Modal } from "@patternfly/react-core";
import CopyToClipboard from "react-copy-to-clipboard";
import ReactDOM from "react-dom";

/**
 * The Page component to handle each element individually.
 */
export default class Page extends Component {
  /**
   * Default constructor for the PageComponent.
   *
   * @param {*} props { pageDef, jsonSchema, pageNumber }
   */
  constructor(props) {
    super(props);
    this.state = {
      elements: [],
      isModalOpen: false,
      resultYaml: ""
    };
    this.handleModalToggle = () => {
      this.setState(({ isModalOpen }) => ({
        isModalOpen: !isModalOpen
      }));
    };
  }

  loadPageChildren() {
    var elements = ElementFactory.newInstances(
      this.props.pageDef.fields,
      this.props.pageDef.buttons,
      this.props.jsonSchema,
      this.props.pageNumber,
      this
    );

    this.setState({
      elements: elements
    });
  }

  /**
   * Adds a new element to the specific position at the Page and re-render the DOM.
   * @param {int} startIndex
   * @param {Element} element
   */
  addElements(startIndex, newElements, objectkey) {
    this.state.elements.forEach((element, i) => {
      if (element.props != undefined && element.props.ids != undefined) {
        if (element.props.ids.fieldGroupId === objectkey) {
          startIndex = startIndex + i;
        }
      }
    });

    if (Array.isArray(newElements)) {
      var elements = this.state.elements;
      newElements.forEach((element, count) => {
        // update the json state
        this.props.pageDef.fields.splice(
          startIndex + count - 1,
          0,
          JSON.parse(JSON.stringify(element.props.fieldDef))
        );
        // add the elements dynamically
        elements.splice(startIndex + count, 0, element);
      });

      this.setState({ elements: elements });
    } else {
      throw new Error(
        "When adding new elements to the page, please use an Array. Got: ",
        newElements
      );
    }
  }

  /**
   * Removes elements from the Page.
   *
   * @param {int} startIndex
   * @param {int} elementCount
   */
  deleteElements(startIndex, elementCount, objectkey) {
    this.state.elements.forEach((element, i) => {
      // console.log(element.page.props.key);
      if (element.props != undefined && element.props.ids != undefined) {
        if (element.props.ids.fieldGroupId === objectkey) {
          startIndex = startIndex + i;
        }
      }
    });
    var elements = this.state.elements;
    this.props.pageDef.fields.splice(startIndex - 1, elementCount);
    elements.splice(startIndex, elementCount);
    this.setState({ elements: elements });
  }

  componentDidMount() {
    this.loadPageChildren();
    this.renderPages();
  }

  getElements() {
    return this.state.elements;
  }
  editYaml() {
    this.createSampleYaml();

    this.handleModalToggle();
    //alert(YAML.safeDump(Dot.object(jsonObject)));
  }
  createResultYaml = jsonObject => {
    var resultYaml = YAML.safeDump(Dot.object(jsonObject));
    this.setState({
      resultYaml
    });

    return resultYaml;
  };
  onChangeYaml = value => {
    this.setResultYaml(value);
  };

  deploy = () => {
    //alert("deploy here");
    this.createSampleYaml();
  };

  setResultYaml = resultYaml => {
    this.setState({
      resultYaml
    });
  };

  createSampleYaml() {
    const jsonObject = {};
    this.state.elements.forEach(element => {
      if (element.props != undefined && element.props.ids != undefined) {
        if (
          element.props.fieldDef.value !== undefined &&
          element.props.fieldDef.value !== ""
        ) {
          let jsonPath = this.getJsonSchemaPathForYaml(
            element.props.fieldDef.jsonPath
          );
          jsonObject[jsonPath] = element.props.fieldDef.value;
        }
      }
    });
    var result = this.createResultYaml(jsonObject);
    console.log(result);
    fetch("/", {
      method: "POST",
      body: JSON.stringify(result),
      headers: {
        "Content-Type": "application/json"
      }
    })
      .then(res => res.json())
      .then(response => console.log("Success:", JSON.stringify(response)))
      .catch(error => console.error("Error:", error));
  }

  getJsonSchemaPathForYaml(jsonPath) {
    //console.log("json Path: " + jsonPath);
    jsonPath = jsonPath.slice(2, jsonPath.length);

    //console.log("jsonSchema Path: " + jsonPath);
    return jsonPath;
  }

  renderPages() {
    //const pages = this.state.jsonForm.pages;
    // console.error("renderPages1");
    /* // const pagesJsx = this.buildPages();
    const wizardJsx= this.buildPages();
  // const steps = this.buildPages();;
    console.error("renderPages2:::" + wizardJsx);

    this.setState({ wizardJsx });
    //this.setState(steps)*/
    var div = document.createElement("div");
    div.id = "footerDiv";
    var footerElem = document.getElementsByTagName("FOOTER");
    //alert("footerElem"+footerElem);
    // var index = document.getElementsByTagName("FOOTER").length;

    // var buttonsJsx = [];
    var buttonJsx = (
      // <ActionGroup fieldid="footer_buttons" key="footer_buttons_key">

      <Button
        variant="primary"
        id="deploy"
        key="deploKey"
        onClick={this.deploy}
      >
        Deploy{" "}
      </Button>
      // </ActionGroup>
    );

    //ReactDOM.render(fieldJsx, footerElem[0]);
    //footerElem.appendChild(fieldJsx);
    // footerElem.innerHTML +=fieldJsx;
    // footerElem.innerHTML = (fieldJsx);
    if (footerElem[0] != undefined) {
      footerElem[0].appendChild(div);
    }
    ReactDOM.render(buttonJsx, document.getElementById("footerDiv"));
  }

  render() {
    const { isModalOpen } = this.state;

    return (
      <Form id={"form-page-" + this.props.pageNumber}>
        <div key={"page" + this.props.pageNumber}>
          {this.state.elements.map(element => {
            return element.getJsx();
          })}

          <Modal
            title=" "
            width={"200%"}
            isOpen={isModalOpen}
            onClose={this.handleModalToggle}
            actions={[
              <CopyToClipboard
                key="yaml_copy"
                className="pf-c-button pf-m-primary"
                onCopy={this.onCopyYaml}
                text={this.state.resultYaml}
              >
                <button key="yaml_button_copy">Copy to clipboard</button>
              </CopyToClipboard>,
              <Button
                key="cancel"
                variant="secondary"
                onClick={this.handleModalToggle}
              >
                Cancel
              </Button>
            ]}
          >
            <TextArea
              id="yaml_edit_text"
              key="yaml_text"
              onChange={this.onChangeYaml}
              rows={35}
              cols={35}
              value={this.state.resultYaml}
            />
          </Modal>
        </div>
      </Form>
    );
  }
}
