import React from "react";
import PropTypes from "prop-types";
import {
  Title,
  Form,
  Text,
  TextContent,
  TextVariants,
  EmptyState,
  EmptyStateVariant,
  EmptyStateIcon,
  EmptyStateBody,
  Expandable
} from "@patternfly/react-core";
import { CheckCircleIcon, ErrorCircleOIcon } from "@patternfly/react-icons";

export default class ReviewPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      showErrorMsg: false
    };
    this.onShowErrorMsg = () => {
      this.setState({
        showErrorMsg: !this.state.showErrorMsg
      });
    };
  }

  render() {
    if (this.props.deployment.deployed === true) {
      if (this.props.deployment.error === undefined) {
        return (
          <EmptyState variant={EmptyStateVariant.full}>
            <EmptyStateIcon icon={CheckCircleIcon} />
            <Title headingLevel="h5" size="lg">
              Application deployed
            </Title>
            <EmptyStateBody>
              The application has been deployed successfully
            </EmptyStateBody>
          </EmptyState>
        );
      } else {
        const { showErrorMsg } = this.state;
        return (
          <EmptyState variant={EmptyStateVariant.full}>
            <EmptyStateIcon icon={ErrorCircleOIcon} />
            <Title headingLevel="h5" size="lg">
              Unable to deploy the application
            </Title>
            <Expandable
              toggleText={showErrorMsg ? "Hide details" : "Show details"}
              onToggle={this.onShowErrorMsg}
              isExpanded={showErrorMsg}
            >
              <TextContent>
                <Text component={TextVariants.small}>
                  {this.props.deployment.error}
                </Text>
              </TextContent>
            </Expandable>
          </EmptyState>
        );
      }
    } else {
      return (
        <Form>
          <Title headingLevel="h1" size="2xl">
            Confirm the installation settings
          </Title>
          <TextContent>
            <Text component={TextVariants.p}>
              Review the information provided and click Deploy to configure your
              project.
              <br />
              Use the Back button to make changes.
            </Text>
          </TextContent>
        </Form>
      );
    }
  }
}

ReviewPage.propTypes = {
  title: PropTypes.string.isRequired,
  deployment: PropTypes.object.isRequired
};
