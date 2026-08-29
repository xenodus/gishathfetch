import { Button, Modal } from "react-bootstrap";
import {
  TELEGRAM_BOT_HANDLE_URL,
  TELEGRAM_BOT_PRIVACY_URL,
} from "../constants";
import LazyMapIframe from "./LazyMapIframe";

const Modals = ({
  showMap,
  onHideMap,
  showFaq,
  onHideFaq,
  showPrivacy,
  onHidePrivacy,
  onShowPrivacy,
  lgsMapData,
}) => {
  return (
    <>
      {/* Map Modal */}
      <Modal show={showMap} onHide={onHideMap} size="xl">
        <Modal.Header closeButton>
          <Modal.Title id="map-list">Where are the shops?</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="mb-4">
            <ul style={{ paddingLeft: "1rem" }}>
              {lgsMapData.map((shop, i) => (
                // biome-ignore lint/suspicious/noArrayIndexKey: Static map data
                <li key={i}>
                  <a href={`#${shop.id}`} className="link-offset-2">
                    {shop.name}
                  </a>
                </li>
              ))}
            </ul>
          </div>
          {lgsMapData.map((shop, i) => (
            // biome-ignore lint/suspicious/noArrayIndexKey: Static map data
            <div id={shop.id} key={i} className="mb-4 map-item">
              <h5>{shop.name}</h5>
              <div className="mb-2">{shop.address}</div>
              <div className="mb-2">
                <a href={shop.website} target="_blank" rel="noreferrer">
                  {shop.website}
                </a>
              </div>
              <LazyMapIframe
                src={shop.iframe}
                title={shop.name}
                isActive={showMap}
              />
              <div>
                <Button
                  variant="primary"
                  onClick={() =>
                    document.getElementById("map-list").scrollIntoView()
                  }
                >
                  Back to top
                </Button>
                <Button
                  variant="secondary"
                  className="ms-2"
                  onClick={onHideMap}
                >
                  Close
                </Button>
              </div>
            </div>
          ))}
        </Modal.Body>
        <Modal.Footer className="justify-content-start">
          &copy; 2023 gishathfetch.com by{" "}
          <a href="https://github.com/xenodus" target="_blank" rel="noreferrer">
            xenodus
          </a>{" "}
          |{" "}
          <Button
            variant="link"
            className="p-0 text-decoration-none"
            onClick={onShowPrivacy}
          >
            privacy policy
          </Button>
        </Modal.Footer>
      </Modal>

      {/* FAQ Modal */}
      <Modal show={showFaq} onHide={onHideFaq} size="xl">
        <Modal.Header
          closeButton
          className="border-bottom border-dark border-opacity-25"
        >
          <Modal.Title id="faq-list">FAQs</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="mb-4">
            <ol style={{ paddingLeft: "1rem" }}>
              <li>
                <a href="#faq-q1" className="link-offset-2">
                  How does Gishath Fetch work?
                </a>
              </li>
              <li>
                <a href="#faq-q2" className="link-offset-2">
                  Is Gishath Fetch free to use?
                </a>
              </li>
              <li>
                <a href="#faq-q3" className="link-offset-2">
                  How do I get in touch?
                </a>
              </li>
              <li>
                <a href="#faq-q4" className="link-offset-2">
                  Why aren't all results shown?
                </a>
              </li>
              <li>
                <a href="#faq-q5" className="link-offset-2">
                  Why do some store links show the wrong condition?
                </a>
              </li>
              <li>
                <a href="#faq-q6" className="link-offset-2">
                  Why do searches take longer with more stores selected?
                </a>
              </li>
              <li>
                <a href="#faq-q7" className="link-offset-2">
                  How can I get accurate results from every store?
                </a>
              </li>
              <li>
                <a href="#faq-q8" className="link-offset-2">
                  What is in Trending?
                </a>
              </li>
              <li>
                <a href="#faq-q9" className="link-offset-2">
                  Where else can I find Gishath Fetch?
                </a>
              </li>
            </ol>
          </div>

          <div className="mb-4" id="faq-q1">
            <div className="q-header">
              <h5>1. How does Gishath Fetch work?</h5>
            </div>
            <div className="q-answer">
              <p>
                Gishath Fetch searches the local game stores (LGS) you have
                selected at the same time. It combines their listings, filters
                for better matches, and sorts everything by price so you can
                compare quickly.
              </p>
            </div>
          </div>
          <div className="mb-4" id="faq-q2">
            <div className="q-header">
              <h5>2. Is Gishath Fetch free to use?</h5>
            </div>
            <div className="q-answer">
              <p>
                Yes. Gishath Fetch is a passion project for fellow MTG players,
                and there are no plans to put it behind a paywall.
              </p>
              <p>
                Ads help cover operating costs. If you have feedback on ad
                placements, get in touch (below).
              </p>
              <p>
                If you want to support the site directly, you can do so on{" "}
                <a
                  href="https://www.patreon.com/GishathFetch"
                  target="_blank"
                  rel="noreferrer"
                >
                  Patreon
                </a>
                . Free Patreon membership is also available if you just want to
                follow updates.
              </p>
            </div>
          </div>
          <div className="mb-4" id="faq-q3">
            <div className="q-header">
              <h5>3. How do I get in touch?</h5>
            </div>
            <div className="q-answer">
              <p>
                Have a suggestion, found a bug, or just want to say hello? Email{" "}
                <a
                  href="mailto:contact@alvinyeoh.com"
                  target="_blank"
                  rel="noreferrer"
                >
                  contact@alvinyeoh.com
                </a>
                .
              </p>
            </div>
          </div>
          <div className="mb-4" id="faq-q4">
            <div className="q-header">
              <h5>4. Why aren't all results shown?</h5>
            </div>
            <div className="q-answer">
              <p>
                Gishath Fetch does not paginate store search results. For most
                LGSs it only pulls the first results page, or about the first 25
                products, then filters and sorts what it gets.
              </p>
              <p>
                That is usually enough for a specific card search, because the
                best matches tend to appear first. Broad queries (for example
                basic lands or a short name with many printings) can spill onto
                later pages on the store's own site — check the LGS directly if
                you need every printing or condition.
              </p>
              <p>
                Out-of-stock listings are also left out on purpose, so a store
                may have a printing that does not appear here.
              </p>
            </div>
          </div>
          <div className="mb-4" id="faq-q5">
            <div className="q-header">
              <h5>5. Why do some store links show the wrong condition?</h5>
            </div>
            <div className="q-answer">
              <p>
                Some store product links open on the wrong condition (for
                example Lightly Played instead of Near Mint). Toggle the
                condition on the store's site to see the correct listing. That
                comes from the store's website, not Gishath Fetch.
              </p>
            </div>
          </div>
          <div className="mb-4" id="faq-q6">
            <div className="q-header">
              <h5>6. Why do searches take longer with more stores selected?</h5>
            </div>
            <div className="q-answer">
              <p>
                Each extra store adds work. We throttle requests so we stay
                within store rate limits, so more shops means more time waiting
                in the queue. Selecting fewer stores makes searches faster, is
                gentler on the shops, and helps keep operating costs down.
              </p>
              <p>
                <strong>Agora Hobby</strong> and{" "}
                <strong>Mox &amp; Lotus</strong> are generally slower to respond
                than other stores, so searches that include them often take
                longer and time out more frequently. If you need faster results,
                try searching without those two selected.
              </p>
            </div>
          </div>
          <div className="mb-4" id="faq-q7">
            <div className="q-header">
              <h5>7. How can I get accurate results from every store?</h5>
            </div>
            <div className="q-answer">
              <p>
                Some stores need a full card name to search reliably. A short or
                partial name may return nothing from some shops even when others
                find it.
              </p>
              <p>
                Use the <strong>auto-suggest</strong> to pick a card, or type
                the <strong>full card name</strong>. That is the most reliable
                way to get consistent results across the stores you selected.
              </p>
            </div>
          </div>
          <div className="mb-4" id="faq-q8">
            <div className="q-header">
              <h5>8. What is in Trending?</h5>
            </div>
            <div className="q-answer">
              <p>
                <strong>Trending</strong> shows the most searched card names on
                Gishath Fetch for the time range you pick (24 hours, 30 days, 6
                months, or 1 year). Tap a name to run that search.
              </p>
              <p>
                When available, you can also switch to{" "}
                <strong>Top risers @ CK (24h)</strong> or{" "}
                <strong>Top drops @ CK (24h)</strong> to see Card Kingdom cards
                with the largest USD price increases or decreases over the last
                24 hours.
              </p>
            </div>
          </div>
          <div className="mb-4" id="faq-q9">
            <div className="q-header">
              <h5>9. Where else can I find Gishath Fetch?</h5>
            </div>
            <div className="q-answer">
              <p>
                You can also use Gishath Fetch on Telegram. Open{" "}
                <a
                  href={TELEGRAM_BOT_HANDLE_URL}
                  target="_blank"
                  rel="noreferrer"
                >
                  @GishathFetchBot
                </a>{" "}
                and send <strong>/price</strong> followed by a card name for the
                cheapest in-stock match across stores, or <strong>/ck</strong>{" "}
                for a Card Kingdom price lookup. Send <strong>/help</strong> in
                the chat for usage examples.
              </p>
              <p>
                The bot links back to this site for the full search experience.
                For what the bot collects and how that data is used, see the{" "}
                <a
                  href={TELEGRAM_BOT_PRIVACY_URL}
                  target="_blank"
                  rel="noreferrer"
                >
                  Telegram bot privacy policy
                </a>
                .
              </p>
            </div>
          </div>
        </Modal.Body>
        <Modal.Footer className="justify-content-start">
          &copy; 2023 gishathfetch.com by{" "}
          <a href="https://github.com/xenodus" target="_blank" rel="noreferrer">
            xenodus
          </a>{" "}
          |{" "}
          <Button
            variant="link"
            className="p-0 text-decoration-none"
            onClick={onShowPrivacy}
          >
            privacy policy
          </Button>
        </Modal.Footer>
      </Modal>

      {/* Privacy Modal */}
      <Modal show={showPrivacy} onHide={onHidePrivacy} size="xl">
        <Modal.Header
          closeButton
          className="border-bottom border-dark border-opacity-25"
        >
          <Modal.Title>Privacy Policy</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div>
            <p className="fw-bold">Access Logs</p>
            <p>
              This website collects personal data through its server access
              logs. When you access this website, your internet address is
              automatically collected and placed in our access logs. We record
              the URLs of the pages you visit, the times and dates of such
              visits.
            </p>
            <p>
              This information may include Internet protocol (IP) addresses,
              browser type and version, internet service provider (ISP),
              referring/exit pages, operating system, date/time stamp, and/or
              clickstream data, number of visits, websites from which you
              accessed our site (Referrer), and websites that are accessed by
              your system via our website.
            </p>
            <p>
              The processing of this data is necessary for the provision and the
              security of this website.
            </p>
          </div>
          <div>
            <p className="fw-bold">Google Analytics</p>
            <p>
              This website uses Google Analytics. Google Analytics employs
              cookies that are stored on your computer to facilitate an analysis
              of your use of the website. The information generated by these
              cookies, such as time, place and frequency of your visits to our
              site, including your IP address, is transmitted to Google.
            </p>
            <p>
              Google Analytics offers a deactivation add-on for most current
              browsers that provides you with more control over what data Google
              can collect on websites you access. You can find additional
              information about the add-on here.
            </p>
          </div>
        </Modal.Body>
        <Modal.Footer className="justify-content-start">
          &copy; 2023 gishathfetch.com by{" "}
          <a href="https://github.com/xenodus" target="_blank" rel="noreferrer">
            xenodus
          </a>
        </Modal.Footer>
      </Modal>
    </>
  );
};

export default Modals;
