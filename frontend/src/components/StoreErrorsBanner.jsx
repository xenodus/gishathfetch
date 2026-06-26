import { Alert } from "react-bootstrap";
import { formatStoreErrorsSummary } from "../utils/storeErrors";

const StoreErrorsBanner = ({ storeErrors, onDismiss }) => {
  if (!storeErrors?.length) {
    return null;
  }

  return (
    <Alert
      variant="warning"
      className="mb-3"
      role="status"
      aria-live="polite"
      dismissible={!!onDismiss}
      onClose={onDismiss}
    >
      {formatStoreErrorsSummary(storeErrors)}
    </Alert>
  );
};

export default StoreErrorsBanner;
