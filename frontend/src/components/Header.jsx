import { Moon, Sun } from "react-feather";
import { LGS_OPTIONS, SITE_TAGLINE } from "../constants";

const Header = ({ theme, onToggleTheme }) => {
  const isDarkMode = theme === "dark";

  return (
    <div className="mb-3 text-center">
      <div className="d-flex flex-row align-items-center justify-content-center mb-1 position-relative">
        <div className="position-absolute top-0 end-0">
          <button
            type="button"
            className="btn btn-sm btn-outline-primary theme-toggle-btn"
            onClick={onToggleTheme}
            aria-label={`Switch to ${isDarkMode ? "light" : "dark"} mode`}
            title={`Switch to ${isDarkMode ? "light" : "dark"} mode`}
            aria-pressed={isDarkMode}
          >
            {isDarkMode ? (
              <Sun size={16} aria-hidden="true" />
            ) : (
              <Moon size={16} aria-hidden="true" />
            )}
          </button>
        </div>
        <div>
          <a href="/">
            <img
              id="logo"
              src="img/gishath-fetch-web.png"
              className="mb-2"
              alt="Gishath Fetch"
            />
          </a>
        </div>
      </div>
      <div className="px-3">
        <h1 className="fw-medium fs-4">
          - Gishath Fetch -<br />
          {SITE_TAGLINE}
          <br />
          <span className="fs-6 fw-normal">
            Search {LGS_OPTIONS.length} LGS and online shops at once
          </span>
        </h1>
      </div>
    </div>
  );
};

export default Header;
