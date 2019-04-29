import React from "react";
import { Form } from "@patternfly/react-core";

import PageBase from "./PageBase";

export default class EachPage extends PageBase {
  constructor(props) {
    super(props);

    this.state = {
      jsonForm: this.props.jsonForm,
      pageNumber: this.props.pageNumber,
      children: [],
      objectMap: new Map(),
      objectCntMap: new Map()
    };
  }

  componentDidMount() {
    this.renderComponents();
  }

  render() {
    return <Form isHorizontal>{this.state.children}</Form>;
  }
}
