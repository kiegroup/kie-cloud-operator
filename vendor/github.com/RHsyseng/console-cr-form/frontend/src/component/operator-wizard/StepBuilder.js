import React from "react";
import * as jsonLoader from "./FormJsonLoader";
import Page from "./page-component/Page";

export default class StepBuilder {
  constructor() {
    this.objectMap = new Map();
  }

  /**
   * Just a placeholder while we build the actual ones. Wizard complains if we don't have at least one step defined.
   */
  buildPlaceholderStep() {
    return [
      {
        id: 0,
        name: "Loading",
        component: <div>Loading</div>
      }
    ];
  }

  buildSteps() {
    return Promise.all([
      jsonLoader.loadJsonForm,
      jsonLoader.loadJsonSchema
    ]).then(values => {
      this.jsonForm = values[0];
      this.jsonSchema = values[1];
      let steps = [];
      let pageId = 1;
      this.jsonForm.pages.forEach(page => {
        const step = this.buildStep(page, pageId);
        if (Array.isArray(page.subPages) && page.subPages.length > 0) {
          step.steps = [];
          page.subPages.forEach(subPage => {
            step.steps.push(this.buildStep(subPage, pageId));
            pageId++;
          });
        } else {
          pageId++;
        }
        steps.push(step);
      });
      return {
        steps: steps,
        pages: this.jsonForm.pages,
        maxSteps: pageId
      };
    });
  }

  storeObjectMap(key, value) {
    this.objectMap.set(key, value);
  }
  getObjectMap(key) {
    return this.objectMap.get(key);
  }
  removeObjectMapPrefix(prefix) {
    for (const key of this.objectMap.keys()) {
      if (key.startsWith(prefix)) {
        this.objectMap.delete(key);
      }
    }
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
          jsonSchema={this.jsonSchema}
          pageNumber={id}
          pages={this.jsonForm.pages} //TODO: try to remove
          storeObjectMap={this.storeObjectMap}
          getObjectMap={this.getObjectMap}
          removeObjectMapPrefix={this.removeObjectMapPrefix}
          objectMap={this.objectMap}
        />
      )
    };
  }
}
