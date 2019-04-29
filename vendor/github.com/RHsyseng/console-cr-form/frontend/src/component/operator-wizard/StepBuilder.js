import React from "react";
import { FormJsonLoader } from "./FormJsonLoader";
import Page from "./page-component/Page";

export default class StepBuilder {
  constructor() {
    this.loader = new FormJsonLoader({
      elementIdJson: "golang_json_form",
      elementIdJsonSchema: "golang_json_schema"
    });
  }

  /**
   * Just a placeholder while we build the actual ones. Wizard complains if we don't have at least one step defined.
   */
  buildPlaceholderStep() {
    return {
      id: 0,
      name: "Loading",
      component: <div>Loading</div>,
      enableNext: true
    };
  }

  buildSteps() {
    var steps = [];
    var pages = this.loader.jsonForm.pages;
    pages.forEach((page, count) => {
      var step = this.buildStep(page, count + 1);

      if (Array.isArray(page.subPages) && page.subPages.length > 0) {
        step.steps = [];
        page.subPages.forEach((subPage, subPageCount) => {
          step.steps.push(this.buildStep(subPage, subPageCount + 1));
        });
      }

      steps.push(step);
    });

    return steps;
  }

  /**
   * Builds a collection of steps based on the page definitions
   * @param {JSON of page def} pageDefs
   */
  buildStep(pageDef, id) {
    var stepName = "Page " + id;
    if (pageDef.label !== undefined && pageDef.label !== "") {
      stepName = pageDef.label;
    }
    return {
      id: id,
      name: stepName, //TODO: this info could be set on the page def
      component: (
        <Page
          key={"page" + id}
          pageDef={pageDef}
          jsonSchema={this.loader.jsonSchema}
          pageNumber={id}
        />
      ),
      enableNext: true //TODO: need to add logic - will enable next only if all fields are valid
    };
  }
}
