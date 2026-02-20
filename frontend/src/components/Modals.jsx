import { Button, Modal } from "react-bootstrap";

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
              <iframe
                className="w-100 border border-dark mb-3"
                style={{ minHeight: "450px" }}
                src={shop.iframe}
                allowFullScreen=""
                loading="lazy"
                referrerPolicy="no-referrer-when-downgrade"
                title={shop.name}
              ></iframe>
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
                  Known issues
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
                Gishath Fetch searches the selected local game stores' (LGS)
                website concurrently in the background, performs filtering of
                result for higher accuracy and returns the compiled result
                sorted by price.
              </p>
            </div>
          </div>
          <div className="mb-4" id="faq-q2">
            <div className="q-header">
              <h5>2. Is Gishath Fetch free to use?</h5>
            </div>
            <div className="q-answer">
              <p>
                Gishath Fetch is build as a project of passion for fellow MTG
                enthusiasts. There are no plans currently nor in the foreseeable
                future to paywall it.
              </p>
              <p>
                Google ads are being served to hopefully generate sufficient
                earnings to cover the operating cost. This is still being tested
                and if you have any feedback about the ad placements, feel free
                to get in touch (below).
              </p>
              <p>
                If you would like to support Gishath Fetch directly, you can do
                so via this{" "}
                <a
                  href="https://www.patreon.com/GishathFetch"
                  target="_blank"
                  rel="noreferrer"
                >
                  Patreon
                </a>{" "}
                ❤️
              </p>
              <p>
                You can also join as a free member on{" "}
                <a
                  href="https://www.patreon.com/GishathFetch"
                  target="_blank"
                  rel="noreferrer"
                >
                  Patreon
                </a>{" "}
                to follow the latest news from Gishath Fetch.
              </p>
            </div>
          </div>
          <div className="mb-4" id="faq-q3">
            <div className="q-header">
              <h5>3. How do I get in touch?</h5>
            </div>
            <div className="q-answer">
              <p>
                Have a suggestion, want to report a bug or just want to get in
                touch? Drop an email to{" "}
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
                Gishath Fetch currently only returns the result from the first
                page of most LGSs' websites or the first 25 results.
              </p>
              <p>
                This is generally not a problem as the most accurate results
                would be on the initial pages unless it's cards with many
                variations (e.g. basic lands). In such cases, you may want to
                visit the LGSs' websites directly.
              </p>
            </div>
          </div>
          <div className="mb-4" id="faq-q5">
            <div className="q-header">
              <h5>5. Known issues</h5>
            </div>
            <div className="q-answer">
              <p>
                Links to some of the LGSs' card variants (e.g. Lightly Played)
                are not showing the correct item upon landing on the LGS's
                website. You are required to toggle between variants (e.g. click
                Near Mint and then back to Lightly Played) to see the correct
                item. This is a problem with the LGS's website and not Gishath
                Fetch.
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
