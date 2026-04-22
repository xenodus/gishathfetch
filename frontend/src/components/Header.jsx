const Header = ({ theme, onToggleTheme }) => {
  const isDarkMode = theme === "dark";

  return (
    <div className="mb-3 text-center">
      <div className="d-flex flex-row align-items-center justify-content-center mb-1 position-relative">
        <div className="position-absolute top-0 end-0">
          <button
            type="button"
            className="btn btn-sm btn-outline-secondary theme-toggle-btn"
            onClick={onToggleTheme}
            aria-label={`Switch to ${isDarkMode ? "light" : "dark"} mode`}
            title={`Switch to ${isDarkMode ? "light" : "dark"} mode`}
            aria-pressed={isDarkMode}
          >
            <span className="theme-toggle-icon" aria-hidden="true">
              💡
            </span>
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
          Magic: The Gathering Price Checker for Singapore's LGS
        </h1>
      </div>
    </div>
  );
};

export default Header;
