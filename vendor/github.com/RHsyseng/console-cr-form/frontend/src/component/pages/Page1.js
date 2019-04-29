import PageBase from "../PageBase";

export default class Page1 extends PageBase {
  constructor(props) {
    super(props);

    this.state = {
      jsonForm: this.props.jsonForm,
      children: [],
      pageNumber: 0,
      objectMap: new Map(),
      objectCntMap: new Map()
    };
  }
}
