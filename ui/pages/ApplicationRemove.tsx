import { CircularProgress } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import Alert from "../components/Alert";
import AuthAlert from "../components/AuthAlert";
import Button from "../components/Button";
import Flex from "../components/Flex";
import Page from "../components/Page";
import Spacer from "../components/Spacer";
import Text from "../components/Text";
import { AppContext } from "../contexts/AppContext";
import { useAppRemove } from "../hooks/applications";
import { GrpcErrorCodes } from "../lib/types";
import { poller } from "../lib/utils";

type Props = {
  className?: string;
  name: string;
};

const RepoRemoveStatus = ({ done }: { done: boolean }) =>
  done ? (
    <Alert
      severity="info"
      title="Removed from Git Repo"
      message="The application successfully removed from your git repository"
    />
  ) : (
    <CircularProgress />
  );

const ClusterRemoveStatus = ({ done }: { done: boolean }) =>
  done ? (
    <Alert
      severity="success"
      title="Removed from cluster"
      message="The application was removed from your cluster"
    />
  ) : (
    <CircularProgress />
  );

const Prompt = ({ onRemove, name }: { name: string; onRemove: () => void }) => (
  <Flex column center>
    <Flex wide center>
      <Text size="large" bold>
        Are you sure you want to remove the application {name}?
      </Text>
    </Flex>
    <Flex wide center>
      <Spacer padding="small">
        Removing this application will remove any Kubernetes objects that were
        created by the application
      </Spacer>
    </Flex>
    <Flex wide center>
      <Spacer padding="small">
        <Button onClick={onRemove} variant="contained" color="secondary">
          Remove {name}
        </Button>
      </Spacer>
    </Flex>
  </Flex>
);

function ApplicationRemove({ className, name }: Props) {
  const { applicationsClient } = React.useContext(AppContext);

  const [removeRes, removing, error, remove] = useAppRemove();
  const [provider, setProvider] = React.useState(null);
  const [removeComplete, setRemoveComplete] = React.useState(false);

  React.useEffect(() => {
    (async () => {
      const {
        application: { url },
      } = await applicationsClient.GetApplication({
        name,
        namespace: "wego-system",
      });

      const { provider } = await applicationsClient.ParseRepoURL({ url });

      setProvider(provider);
    })();
  }, [name]);

  React.useEffect(() => {
    if (removing || error) {
      return;
    }

    const poll = poller(() => {
      applicationsClient
        .GetApplication({ name, namespace: "wego-system" })
        .catch(() => {
          // Once we get a 404, the app is gone for good
          clearInterval(poll);
          setRemoveComplete(true);
        });
    }, 5000);

    return () => {
      clearInterval(poll);
    };
  }, [removeRes]);

  console.log(error);

  const handleRemoveClick = () => {
    remove(provider, { name, namespace: "wego-system" });
  };

  if (!provider) {
    return <CircularProgress />;
  }
  return (
    <Page className={className}>
      {error &&
        (error.code === GrpcErrorCodes.Unauthenticated ? (
          <AuthAlert title="" provider={provider} onClick={null} />
        ) : (
          <Alert title="" message={error?.message} />
        ))}
      {removing && <CircularProgress />}
      {!removing && <Prompt name={name} onRemove={handleRemoveClick} />}
      {(removing || removeRes) && <RepoRemoveStatus done={!removing} />}
      {!removeComplete && removeRes && (
        <ClusterRemoveStatus done={!removeComplete} />
      )}
    </Page>
  );
}

export default styled(ApplicationRemove).attrs({
  className: ApplicationRemove.name,
})``;
