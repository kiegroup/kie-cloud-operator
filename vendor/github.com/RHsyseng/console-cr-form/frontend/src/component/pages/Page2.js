import PageBase from "../PageBase";

export default class Page2 extends PageBase {
  constructor(props) {
    super(props);

    this.state = {
      jsonForm: this.props.jsonForm,
      children: [],
      pageNumber: 1,
      objectMap: new Map(),
      objectCntMap: new Map(),
      ssoORldap: ""
    };
  }
}
