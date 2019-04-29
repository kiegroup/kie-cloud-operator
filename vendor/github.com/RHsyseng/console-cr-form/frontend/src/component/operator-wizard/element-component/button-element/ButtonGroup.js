import React from "react";
import { ActionGroup } from "@patternfly/react-core";
import { ButtonElement } from "./ButtonElement";

export class ButtonGroup {
  /**
   * Generates an array of child elements based on the button definitions.
   * @param {*} buttonDefs the button definitions
   * @param {int} pageNumber the page numbe￼
r
   * @param {Page} page the Page reference
   */
  constructor(buttonDefs, pageNumber, page) {
    this.buttonDefs = buttonDefs;
    this.pageNumber = pageNumber;
    this.page = page;
  }

  getJsx() {
    var buttonsJsx = [];
    if (
      this.buttonDefs !== undefined &&
      this.buttonDefs !== null &&
      this.buttonDefs !== ""
    ) {
      this.buttonDefs.forEach((buttonDef, i) => {
        var buttonElement = new ButtonElement({
          buttonDef: buttonDef,
          pageNumber: this.pageNumber,
          buttonId: i,
          page: this.page
        });
        buttonsJsx.push(buttonElement.getJsx());
      });
    }

    if (buttonsJsx.length > 0) {
      const actionGroupKey = this.pageNumber + "-action-group";
      return <ActionGroup key={actionGroupKey}>{buttonsJsx}</ActionGroup>;
    }

    return [];
  }
}
