import type { FC } from "react";

import { useNavigation } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { ReportDashboard } from "./components/ReportDashboard";
import { ReportMissingState } from "./components/ReportMissingState";

interface ReportScreenProps {
  route: Route;
}

/** reportId is the only locator; all status and display context comes from the API. */
export const ReportScreen: FC<ReportScreenProps> = ({ route }) => {
  const { navigate } = useNavigation();
  if (!route.params.reportId) {
    return (
      <ReportMissingState
        onBackToWorkspace={() => navigate({ name: "workspace", params: {} })}
      />
    );
  }
  return <ReportDashboard reportId={route.params.reportId} />;
};
