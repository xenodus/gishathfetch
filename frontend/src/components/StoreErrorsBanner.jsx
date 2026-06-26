import { useState } from "react";
import { Alert, Button, Collapse } from "react-bootstrap";
import {
  formatStoreErrorDetail,
  formatStoreErrorsSummary,
} from "../utils/storeErrors";

const StoreErrorsBanner = ({ storeErrors, onDismiss }) => {
  const [showDetails, setShowDetails] = useState(false);

  if (!storeErrors?.length) {
    return null;
  }

  const summary = formatStoreErrorsSummary(storeErrors);

  return (
    <Alert
      variant="warning"
      className="mb-3"
      role="status"
      aria-live="polite"
      dismissible={!!onDismiss}
      onClose={onDismiss}
    >
      <div className="d-flex flex-wrap align-items-center gap-2">
        <span>{summary}</span>
        <Button
          type="button"
          variant="link"
          size="sm"
          className="p-0 align-baseline"
          onClick={() => setShowDetails((current) => !current)}
          aria-expanded={showDetails}
          aria-controls="store-errors-details"
        >
          {showDetails ? "Hide details" : "Show details"}
        </Button>
      </div>
      <Collapse in={showDetails}>
        <div id="store-errors-details" className="mt-2">
          <ul className="mb-0 small ps-3">
            {storeErrors.map((entry) => (
              <li key={entry.store}>{formatStoreErrorDetail(entry)}</li>
            ))}
          </ul>
        </div>
      </Collapse>
    </Alert>
  );
};

export default StoreErrorsBanner;
