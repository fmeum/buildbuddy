import React from 'react';
import Long from 'long';
import moment from 'moment';

import rpcService from '../service/rpc_service'

import InvocationModel from './invocation_model'

import InvocationLoadingComponent from './invocation_loading'
import InvocationInProgressComponent from './invocation_in_progress'
import InvocationNotFoundComponent from './invocation_not_found'

import InvocationOverviewComponent from './invocation_overview'
import InvocationTabsComponent from './invocation_tabs';
import BuildLogsCardComponent from './invocation_build_logs_card'
import ErrorCardComponent from './invocation_error_card';
import InvocationDetailsCardComponent from './invocation_details_card'
import ArtifactsCardComponent from './invocation_artifacts_card'
import RawLogsCardComponent from './invocation_raw_logs_card'
import TargetsCardComponent from './invocation_targets_card'

import DenseInvocationOverviewComponent from './dense/dense_invocation_overview'
import DenseInvocationTabsComponent from './dense/dense_invocation_tabs'


import { invocation } from '../../proto/invocation_ts_proto';

interface State {
  loading: boolean,
  inProgress: boolean,
  notFound: boolean,

  model: InvocationModel,
}

interface Props {
  invocationId: string,
  hash: string
  denseMode: boolean,
}

const largePageSize = 100;
const smallPageSize = 10;

export default class InvocationComponent extends React.Component {
  state: State = {
    loading: true,
    inProgress: false,
    notFound: false,

    model: new InvocationModel(),
  };

  props: Props;

  componentWillMount() {
    document.title = `Invocation ${this.props.invocationId} | Buildbuddy`;
    // TODO(siggisim): Move moment configuration elsewhere
    moment.relativeTimeThreshold('ss', 0);

    this.fetchInvocation();
  }

  fetchInvocation() {
    let request = new invocation.GetInvocationRequest();
    request.query = new invocation.InvocationQuery();
    request.query.invocationId = this.props.invocationId;
    rpcService.service.getInvocation(request).then((response) => {
      console.log(response);

      var showInProgressScreen = false;
      if (response.invocation.length && response.invocation[0].invocationStatus ==
        invocation.Invocation.InvocationStatus.PARTIAL_INVOCATION_STATUS) {
        showInProgressScreen = response.invocation[0].event.length == 0;
        this.fetchUpdatedProgress();
      }

      this.setState({
        inProgress: showInProgressScreen,
        model: InvocationModel.modelFromInvocations(response.invocation as invocation.Invocation[]),
        loading: false
      });
      document.title = `${this.state.model.getUser()}'s ${this.state.model.getCommand()} ${this.state.model.getPattern()} | Buildbuddy`;
    }).catch((error: any) => {
      this.setState({
        notFound: true,
        loading: false
      });
    });
  }

  fetchUpdatedProgress() {
    // Refetch invocation data in 3 seconds to update status.
    setTimeout(() => {
      this.fetchInvocation();
    }, 3000);
  }

  render() {
    if (this.state.loading) {
      return <InvocationLoadingComponent invocationId={this.props.invocationId} />;
    }

    if (this.state.notFound) {
      return <InvocationNotFoundComponent invocationId={this.props.invocationId} />;
    }

    if (this.state.inProgress) {
      return <InvocationInProgressComponent invocationId={this.props.invocationId} />;
    }

    var showAll = !this.props.hash && !this.props.denseMode;

    return (
      <div className={this.props.denseMode ? 'dense' : ''}>
        <div className="shelf">
          {this.props.denseMode ?
            <DenseInvocationOverviewComponent invocationId={this.props.invocationId} model={this.state.model} /> :
            <InvocationOverviewComponent invocationId={this.props.invocationId} model={this.state.model} />
          }

        </div>
        <div className="container">
          {this.props.denseMode ?
            <DenseInvocationTabsComponent hash={this.props.hash} /> :
            <InvocationTabsComponent hash={this.props.hash} />
          }

          {(showAll || this.props.hash == "#log") &&
            <BuildLogsCardComponent model={this.state.model} expanded={this.props.hash == "#log"} />}

          {(showAll || this.props.hash == "#log") &&
            this.state.model.aborted?.aborted.description &&
            <ErrorCardComponent model={this.state.model} />}

          {(!this.props.hash || this.props.hash == "#targets") && !!this.state.model.failed.length &&
            <TargetsCardComponent
              buildEvents={this.state.model.failed}
              iconPath="/image/x-circle.svg"
              presentVerb="failing"
              pastVerb="failed"
              model={this.state.model}
              pageSize={showAll ? smallPageSize : largePageSize}
            />}

          {(!this.props.hash || this.props.hash == "#targets") && !!this.state.model.broken.length &&
            <TargetsCardComponent
              buildEvents={this.state.model.broken}
              iconPath="/image/x-circle.svg"
              presentVerb="broken"
              pastVerb="broken"
              model={this.state.model}
              pageSize={showAll ? smallPageSize : largePageSize}
            />}

          {(!this.props.hash || this.props.hash == "#targets") && !!this.state.model.flaky.length &&
            <TargetsCardComponent
              buildEvents={this.state.model.flaky}
              iconPath="/image/x-circle.svg"
              presentVerb="flaky"
              pastVerb="flaky"
              model={this.state.model}
              pageSize={showAll ? smallPageSize : largePageSize}
            />}

          {(!this.props.hash || this.props.hash == "#targets") && !!this.state.model.succeeded.length &&
            <TargetsCardComponent
              buildEvents={this.state.model.succeeded}
              iconPath="/image/check-circle.svg"
              presentVerb="passing"
              pastVerb="passed"
              model={this.state.model}
              pageSize={showAll ? smallPageSize : largePageSize}
            />}

          {(showAll || this.props.hash == "#details") &&
            <InvocationDetailsCardComponent model={this.state.model} limitResults={!this.props.hash} />}

          {(showAll || this.props.hash == "#artifacts") &&
            <ArtifactsCardComponent model={this.state.model} pageSize={this.props.hash ? largePageSize : smallPageSize} />}

          {(this.props.hash == "#raw") && <RawLogsCardComponent model={this.state.model} pageSize={largePageSize} />}
        </div>
      </div>
    );
  }
}
