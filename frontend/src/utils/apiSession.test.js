import assert from "node:assert/strict";
import {
  DEFAULT_MAINTENANCE_MESSAGE,
  parseMaintenanceFromSessionResponse,
} from "./apiSession.js";

assert.deepEqual(parseMaintenanceFromSessionResponse(new Response(null, { status: 204 })), {
  maintenanceMode: false,
  maintenanceMessage: "",
});

assert.deepEqual(
  parseMaintenanceFromSessionResponse(
    new Response(null, {
      status: 204,
      headers: {
        "X-Maintenance-Mode": "1",
        "X-Maintenance-Message": "Back in 30 minutes.",
      },
    }),
  ),
  {
    maintenanceMode: true,
    maintenanceMessage: "Back in 30 minutes.",
  },
);

assert.deepEqual(
  parseMaintenanceFromSessionResponse(
    new Response(null, {
      status: 204,
      headers: {
        "X-Maintenance-Mode": "1",
      },
    }),
  ),
  {
    maintenanceMode: true,
    maintenanceMessage: DEFAULT_MAINTENANCE_MESSAGE,
  },
);
