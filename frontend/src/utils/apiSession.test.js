import assert from "node:assert/strict";
import {
  DEFAULT_MAINTENANCE_MESSAGE,
  parseMaintenanceFromSessionResponse,
  parseNoticeFromSessionResponse,
  parseSiteStatusFromSessionResponse,
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

assert.deepEqual(parseNoticeFromSessionResponse(new Response(null, { status: 204 })), {
  noticeMessage: "",
});

assert.deepEqual(
  parseNoticeFromSessionResponse(
    new Response(null, {
      status: 204,
      headers: {
        "X-Notice-Message": "  Card Kingdom prices may be delayed.  ",
      },
    }),
  ),
  {
    noticeMessage: "Card Kingdom prices may be delayed.",
  },
);

assert.deepEqual(
  parseSiteStatusFromSessionResponse(
    new Response(null, {
      status: 204,
      headers: {
        "X-Notice-Message": "Welcome to the new season.",
        "X-Maintenance-Mode": "1",
        "X-Maintenance-Message": "Back soon.",
      },
    }),
  ),
  {
    maintenanceMode: true,
    maintenanceMessage: "Back soon.",
    noticeMessage: "Welcome to the new season.",
  },
);
