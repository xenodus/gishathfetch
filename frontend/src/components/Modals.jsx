import React from 'react';
import { Modal, Button } from 'react-bootstrap';

const Modals = ({
    showMap,
    onHideMap,
    showFaq,
    onHideFaq,
    showPrivacy,
    onHidePrivacy,
    onShowPrivacy,
    lgsMapData
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
                        <ul style={{ paddingLeft: '1rem' }}>
                            {lgsMapData.map((shop, i) => (
                                <li key={i}><a href={`#${shop.id}`} className="link-offset-2">{shop.name}</a></li>
                            ))}
                        </ul>
                    </div>
                    {lgsMapData.map((shop, i) => (
                        <div id={shop.id} key={i} className="mb-4 map-item">
                            <h5>{shop.name}</h5>
                            <div className="mb-2">{shop.address}</div>
                            <div className="mb-2"><a href={shop.website} target="_blank" rel="noreferrer">{shop.website}</a></div>
                            <iframe
                                className="w-100 border border-dark mb-3"
                                style={{ minHeight: '450px' }}
                                src={shop.iframe}
                                allowFullScreen=""
                                loading="lazy"
                                referrerPolicy="no-referrer-when-downgrade"
                                title={shop.name}
                            ></iframe>
                            <div>
                                <Button variant="primary" onClick={() => document.getElementById('map-list').scrollIntoView()}>Back to top</Button>
                                <Button variant="secondary" className="ms-2" onClick={onHideMap}>Close</Button>
                            </div>
                        </div>
                    ))}
                </Modal.Body>
                <Modal.Footer className="justify-content-start">
                    © 2023 gishathfetch.com by <a href="https://github.com/xenodus" target="_blank" rel="noreferrer">xenodus</a> | <Button variant="link" className="p-0" onClick={onShowPrivacy}>privacy policy</Button>
                </Modal.Footer>
            </Modal>

            {/* FAQ Modal */}
            <Modal show={showFaq} onHide={onHideFaq} size="xl">
                <Modal.Header closeButton className="border-bottom border-dark border-opacity-25">
                    <Modal.Title id="faq-list">FAQs</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <div className="mb-4">
                        <ol style={{ paddingLeft: '1rem' }}>
                            <li><a href="#faq-q1" className="link-offset-2">How does Gishath Fetch work?</a></li>
                            <li><a href="#faq-q2" className="link-offset-2">Is Gishath Fetch free to use?</a></li>
                            <li><a href="#faq-q3" className="link-offset-2">How do I get in touch?</a></li>
                            <li><a href="#faq-q4" className="link-offset-2">Why aren't all results shown?</a></li>
                            <li><a href="#faq-q5" className="link-offset-2">Known issues</a></li>
                        </ol>
                    </div>

                    <div className="mb-4" id="faq-q1">
                        <h5>1. How does Gishath Fetch work?</h5>
                        <p>Gishath Fetch searches the selected local game stores' (LGS) website concurrently in the background, performs filtering of result for higher accuracy and returns the compiled result sorted by price.</p>
                    </div>
                    <div className="mb-4" id="faq-q2">
                        <h5>2. Is Gishath Fetch free to use?</h5>
                        <p>Gishath Fetch is build as a project of passion for fellow MTG enthusiasts. There are no plans currently nor in the foreseeable future to paywall it.</p>
                        <p>Google ads are being served to hopefully generate sufficient earnings to cover the operating cost. This is still being tested and if you have any feedback about the ad placements, feel free to get in touch (below).</p>
                        <p>If you would like to support Gishath Fetch directly, you can do so via this <a href="https://www.patreon.com/GishathFetch" target="_blank" rel="noreferrer">Patreon</a> ❤️</p>
                    </div>
                    <div className="mb-4" id="faq-q3">
                        <h5>3. How do I get in touch?</h5>
                        <p>Have a suggestion, want to report a bug or just want to get in touch? Drop an email to <a href="mailto:contact@alvinyeoh.com" target="_blank" rel="noreferrer">contact@alvinyeoh.com</a>.</p>
                    </div>
                    <div className="mb-4" id="faq-q4">
                        <h5>4. Why aren't all results shown?</h5>
                        <p>Gishath Fetch currently only returns the result from the first page of most LGSs' websites or the first 25 results.</p>
                    </div>
                    <div className="mb-4" id="faq-q5">
                        <h5>5. Known issues</h5>
                        <p>Links to some of the LGSs' card variants (e.g. Lightly Played) are not showing the correct item upon landing on the LGS's website.</p>
                    </div>
                </Modal.Body>
                <Modal.Footer className="justify-content-start">
                    © 2023 gishathfetch.com by <a href="https://github.com/xenodus" target="_blank" rel="noreferrer">xenodus</a> | <Button variant="link" className="p-0" onClick={onShowPrivacy}>privacy policy</Button>
                </Modal.Footer>
            </Modal>

            {/* Privacy Modal */}
            <Modal show={showPrivacy} onHide={onHidePrivacy} size="xl">
                <Modal.Header closeButton className="border-bottom border-dark border-opacity-25">
                    <Modal.Title>Privacy Policy</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <p className="fw-bold">Access Logs</p>
                    <p>This website collects personal data through its server access logs...</p>
                    <p className="fw-bold">Google Analytics</p>
                    <p>This website uses Google Analytics. Google Analytics employs cookies...</p>
                </Modal.Body>
                <Modal.Footer className="justify-content-start">
                    © 2023 gishathfetch.com by <a href="https://github.com/xenodus" target="_blank" rel="noreferrer">xenodus</a>
                </Modal.Footer>
            </Modal>
        </>
    );
};

export default Modals;
