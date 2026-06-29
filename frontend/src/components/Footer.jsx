import { ArrowUp, FolderPlus, HelpCircle, Map as MapIcon } from "react-feather";
import useBottomChromeLayout from "../hooks/useBottomChromeLayout";
import AdComponent from "./AdComponent";

const scrollToTop = () => {
  window.scrollTo({ top: 0, behavior: "smooth" });
};

const Footer = ({ cartCount, onShowCart, onShowMap, onShowFaq }) => {
  useBottomChromeLayout();

  return (
    <>
      <div className="mt-4 site-footer-spacer">
        <AdComponent />
      </div>

      {/* Fixed Bottom Navigation */}
      <div className="site-bottom-nav bg-primary text-light text-center">
        <div className="d-flex flex-row align-items-center justify-content-center">
          <button
            type="button"
            aria-label={`View saved cards${cartCount > 0 ? ` (${cartCount} items)` : ""}`}
            className="btn btn-link py-1 link-light link-offset-2 link-underline-opacity-0 text-decoration-none border-0"
            onClick={onShowCart}
          >
            <div className="px-2 py-1 d-inline-flex align-items-center">
              <span className="bottom-nav-icon">
                <FolderPlus size={14} aria-hidden="true" />
              </span>
              <span>Saved {cartCount > 0 && `(${cartCount})`}</span>
            </div>
          </button>
          <button
            type="button"
            aria-label="View store locations map"
            className="btn btn-link py-1 link-light link-offset-2 link-underline-opacity-0 text-decoration-none border-0"
            onClick={onShowMap}
          >
            <div className="px-2 py-1 d-inline-flex align-items-center">
              <span className="bottom-nav-icon">
                <MapIcon size={14} aria-hidden="true" />
              </span>
              <span>Map</span>
            </div>
          </button>
          <button
            type="button"
            aria-label="View frequently asked questions"
            className="btn btn-link py-1 link-light link-offset-2 link-underline-opacity-0 text-decoration-none border-0"
            onClick={onShowFaq}
          >
            <div className="px-2 py-1 d-inline-flex align-items-center">
              <span className="bottom-nav-icon">
                <HelpCircle size={14} aria-hidden="true" />
              </span>
              <span>FAQs</span>
            </div>
          </button>
          <button
            type="button"
            aria-label="Scroll to top of page"
            className="btn btn-link py-1 link-light link-offset-2 link-underline-opacity-0 text-decoration-none border-0"
            onClick={scrollToTop}
          >
            <div className="px-2 py-1 d-inline-flex align-items-center">
              <span className="bottom-nav-icon">
                <ArrowUp size={14} aria-hidden="true" />
              </span>
              <span>Top</span>
            </div>
          </button>
        </div>
      </div>
    </>
  );
};

export default Footer;
