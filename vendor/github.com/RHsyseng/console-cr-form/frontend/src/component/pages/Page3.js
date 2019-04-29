import PageBase from "../PageBase";

export default class Page3 extends PageBase {
  constructor(props) {
    super(props);

    this.state = {
      jsonForm: this.props.jsonForm,
      children: [],
      pageNumber: 2,
      objectMap: new Map(),
      objectCntMap: new Map()
    };
  }
}
