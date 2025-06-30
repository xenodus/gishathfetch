document.body.innerHTML += `
    <div class="fixed-bottom bg-primary text-light text-center">
        <div class="d-flex flex-row align-items-center justify-content-center">
            <a data-bs-toggle="offcanvas" href="#offcanvasRight" role="button" aria-controls="offcanvasRight" class="py-1 link-light link-offset-2 link-underline-opacity-0">
                <div class="px-3 py-1">
                    <i data-feather="folder-plus" style="width: 14px; margin-right: 4px; position: relative; bottom: 2px;"></i>Saved<span id="cartNo"></span>
                </div>
            </a>        
            <a href="#" data-bs-toggle="modal" data-bs-target="#map-modal" class="py-1 link-light link-offset-2 link-underline-opacity-0">
                <div class="px-3 py-1">
                    <i data-feather="map" style="width: 14px; margin-right: 4px; position: relative; bottom: 1px;"></i>Map
                </div>
            </a>
            <a href="#" data-bs-toggle="modal" data-bs-target="#faq-modal" class="py-1 link-light link-offset-2 link-underline-opacity-0">
                <div class="px-3 py-1">
                    <i data-feather="help-circle" style="width: 14px; margin-right: 3px; position: relative; bottom: 2px;"></i>FAQs
                </div>
            </a>
            <a href="#top" class="py-1 link-light link-offset-2 link-underline-opacity-0">
                <div class="px-3 py-1">
                    <i data-feather="arrow-up" style="width: 14px; margin-right: 3px; position: relative; bottom: 1px;"></i>Top
                </div>
            </a>
        </div>
    </div>
`;

document.body.innerHTML += `
    <div id="map">
        <div class="modal" id="map-modal" tabindex="-1">
            <div class="modal-dialog modal-xl">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 id="map-list" class="modal-title">Where are the shops?</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                    </div>
                    <div class="modal-body"></div>
                    <div class="modal-footer justify-content-start">
                        &copy; 2023 gishathfetch.com by <a href="https://github.com/xenodus" target="_blank">xenodus</a> | <a href="#" data-bs-toggle="modal" data-bs-target="#privacy-modal">privacy policy</a>
                    </div>
                </div>
            </div>
        </div>
    </div>
`;

document.body.innerHTML += `
    <div id="faq">
        <div class="modal" id="faq-modal" tabindex="-1">
            <div class="modal-dialog modal-xl">
                <div class="modal-content">
                    <div class="modal-header border-bottom border-dark border-opacity-25">
                        <h5 id="faq-list" class="modal-title">FAQs</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                    </div>
                    <div class="modal-body">
                        <div class="mb-4">
                            <ol style="padding-left: 1rem;">
                                <li><a href="#faq-q1" class="link-offset-2">How does Gishath Fetch work?</a></li>
                                <li><a href="#faq-q2" class="link-offset-2">Is Gishath Fetch free to use?</a></li>
                                <li><a href="#faq-q3" class="link-offset-2">How do I get in touch?</a></li>
                                <li><a href="#faq-q4" class="link-offset-2">Why aren't all results shown?</a></li>
                                <li><a href="#faq-q5" class="link-offset-2">Known issues</a></li>
                            </ol>
                        </div>
                        <div>
                            <div class="mb-4" id="faq-q1">
                                <div class="q-header"><h5>1. How does Gishath Fetch work?</h5></div>
                                <div class="q-answer">
                                    <p>Gishath Fetch searches the selected local game stores' (LGS) website concurrently in the background, performs filtering of result for higher accuracy and returns the compiled result sorted by price.</p>
                                </div>
                            </div>
                            <div class="mb-4" id="faq-q2">
                                <div class="q-header"><h5>2. Is Gishath Fetch free to use?</h5></div>
                                <div class="q-answer">
                                    <p>Gishath Fetch is build as a project of passion for fellow MTG enthusiasts. There are no plans currently nor in the foreseeable future to paywall it.</p>
                                    <p>Google ads are being served to hopefully generate sufficient earnings to cover the operating cost. This is still being tested and if you have any feedback about the ad placements, feel free to get in touch (below).</p>
                                    <p>If you would like to support Gishath Fetch directly, you can do so via this <a href="https://www.patreon.com/GishathFetch" target="_blank">Patreon</a> ❤️</p>
                                    <p>You can also join as a free member on <a href="https://www.patreon.com/GishathFetch" target="_blank">Patreon</a> to follow the latest news from Gishath Fetch.</p>
                                </div>
                            </div>
                            <div class="mb-4" id="faq-q3">
                                <div class="q-header"><h5>3. How do I get in touch?</h5></div>
                                <div class="q-answer">
                                    <p>Have a suggestion, want to report a bug or just want to get in touch? Drop an email to <a href="mailto:contact@alvinyeoh.com" target="_blank">contact@alvinyeoh.com</a>.</p>
                                </div>
                            </div>
                            <div class="mb-4" id="faq-q4">
                                <div class="q-header"><h5>4. Why aren't all results shown?</h5></div>
                                <div class="q-answer">
                                    <p>
                                        Gishath Fetch currently only returns the result from the first page of most LGSs' websites or the first 25 results.
                                    </p>
                                    <p>
                                        This is generally not a problem as the most accurate results would be on the initial pages unless it's cards with many variations (e.g. basic lands). In such cases, you may want to visit the LGSs' websites directly.
                                    </p>
                                </div>
                            </div>
                            <div class="mb-4" id="faq-q5">
                                <div class="q-header"><h5>5. Known issues</h5></div>
                                <div class="q-answer">
                                    <p>
                                        Links to some of the LGSs' card variants (e.g. Lightly Played) are not showing the correct item upon landing on the LGS's website. You are required to toggle between variants (e.g. click Near Mint and then back to Lightly Played) to see the correct item. This is a problem with the LGS's website and not Gishath Fetch.
                                    </p>
                                </div>
                            </div>                            
                        </div>
                    </div>
                    <div class="modal-footer justify-content-start">
                        &copy; 2023 gishathfetch.com by <a href="https://github.com/xenodus" target="_blank">xenodus</a> | <a href="#" data-bs-toggle="modal" data-bs-target="#privacy-modal">privacy policy</a>
                    </div>
                </div>
            </div>
        </div>
    </div>
`;

document.body.innerHTML += `
    <div id="privacy">
        <div class="modal" id="privacy-modal" tabindex="-1">
            <div class="modal-dialog modal-xl">
                <div class="modal-content">
                    <div class="modal-header border-bottom border-dark border-opacity-25">
                        <h5 class="modal-title">Privacy Policy</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                    </div>
                    <div class="modal-body">
                        <div>
                            <p class="fw-bold">Access Logs</p>
                            <p>This website collects personal data through its server access logs. When you access this website, your internet address is automatically collected and placed in our access logs. We record the URLs of the pages you visit, the times and dates of such visits.</p>
                            <p>This information may include Internet protocol (IP) addresses, browser type and version, internet service provider (ISP), referring/exit pages, operating system, date/time stamp, and/or clickstream data, number of visits, websites from which you accessed our site (Referrer), and websites that are accessed by your system via our website.</p>
                            <p>The processing of this data is necessary for the provision and the security of this website.</p>
                        </div>
                        <div>
                            <p class="fw-bold">Google Analytics</p>
                            <p>This website uses Google Analytics. Google Analytics employs cookies that are stored on your computer to facilitate an analysis of your use of the website. The information generated by these cookies, such as time, place and frequency of your visits to our site, including your IP address, is transmitted to Google.</p>
                            <p>Google Analytics offers a deactivation add-on for most current browsers that provides you with more control over what data Google can collect on websites you access. You can find additional information about the add-on here.</p>
                        </div>
                    </div>
                    <div class="modal-footer justify-content-start">
                        &copy; 2023 gishathfetch.com by <a href="https://github.com/xenodus" target="_blank">xenodus</a>
                    </div>
                </div>
            </div>
        </div>
    </div>
`;

document.body.innerHTML += `
    <div class="offcanvas offcanvas-end" tabindex="-1" id="offcanvasRight" aria-labelledby="offcanvasRightLabel">
        <div class="offcanvas-header">
            <h5 class="offcanvas-title" id="offcanvasRightLabel">Saved Cards</h5>
            <button type="button" class="btn-close" data-bs-dismiss="offcanvas" aria-label="Close"></button>
        </div>
        <div class="offcanvas-body">
            <div class="mb-3">When a card is saved, a snapshot of it from that point in time is taken. If there is any change in its price or availability, it will not be updated automatically.</div>
            <div id="cartContent">
                <div class="row">
                    <div class="col-6">
                        <div class="text-center mb-2">
                          <a href="https://www.flagshipgames.sg/products/sol-ring-commander-2013?_pos=2&amp;_sid=822ba891b&amp;_ss=r" target="_blank">
                            <img src="https://www.flagshipgames.sg/cdn/shop/products/1b59533a-3e38-495d-873e-2f89fbd08494_370x480.jpg?v=1600893402" loading="lazy" class="img-fluid w-100" alt="Sol Ring [Commander 2013]">
                          </a>
                        </div>                    
                    </div>                                                                                                               
                </div>
            </div>
        </div>
    </div>
`;

// For force updates of storage
let version = "1.0";
let cart = [];
let lgsMap = [
    {
        "id": "agora-map",
        "name": "Agora Hobby",
        "address": "French Rd, #05-164 Blk 809, Singapore 200809",
        "iframe": "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.778050505021!2d103.85967687451628!3d1.3084089617085968!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da19c9f7d7f74d%3A0xeaa1a66df7d4bcd6!2sAgora%20Hobby!5e0!3m2!1sen!2ssg!4v1702820213937!5m2!1sen!2ssg",
    },
    {
        "id": "cardboard-crack-games-map",
        "name": "Cardboard Crack Games",
        "address": "Upper Bukit Timah Rd, #03-28 Beauty World Centre, 144, Singapore 588177",
        "iframe": "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.7233430292667!2d103.7736843749657!3d1.3423740986449086!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da1920d676db93%3A0xe7b298b897da7b52!2sCardboard%20Crack%20Games!5e0!3m2!1sen!2ssg!4v1731824736033!5m2!1sen!2ssg",
    },
    {
        "id": "cards-citadel-map",
        "name": "Cards Citadel",
        "address": "464 Crawford Ln, #02-01, Singapore 190464",
        "iframe": "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.783678524258!2d103.85966947451631!3d1.3048646617197366!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da190c9e183751%3A0xa2119a95d1e683f2!2sCards%20Citadel!5e0!3m2!1sen!2ssg!4v1702820792196!5m2!1sen!2ssg",
    },
    {
        "id": "dueller-point-map",
        "name": "Dueller's Point",
        "address": "450 Hougang Ave 10, B1-541, Singapore 530450",
        "iframe": "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.662159756766!2d103.89300967451602!3d1.3793695614811952!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da163eecb250ff%3A0xc7c259e72671dc62!2sDueller&#39;s%20Point!5e0!3m2!1sen!2ssg!4v1702820876967!5m2!1sen!2ssg",
    },
    {
        "id": "flagship-games-map",
        "name": "Flagship Games",
        "address": "214 Bishan St. 23, B1-223, Singapore 570214",
        "iframe": "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d490.0218996351789!2d103.84829838647084!3d1.3574649065942905!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da173ef6ffcc0b%3A0x880386dee363a253!2sFlagship%20Games!5e0!3m2!1sen!2ssg!4v1734958555684!5m2!1sen!2ssg",
    },
    {
        "id": "games-haven-pl-map",
        "name": "Games Haven - Paya Lebar",
        "address": "736 Geylang Rd, Singapore 389647",
        "iframe": "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d63819.358332241325!2d103.79905633083244!3d1.350592080054757!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da1817d10ac901%3A0x2cacb3a0679089a2!2sGames%20Haven!5e0!3m2!1sen!2ssg!4v1702821045126!5m2!1sen!2ssg",
    },
    {
        "id": "games-haven-ct-map",
        "name": "Games Haven - Chinatown",
        "address": "531 Upper Cross St, #02-02, Singapore 050531",
        "iframe": "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d63821.04849133675!2d103.76901934863278!3d1.2846211999999948!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da196e26ea750d%3A0x688a1d11efebe621!2sGames%20Haven%20-%20Chinatown!5e0!3m2!1sen!2ssg!4v1702821183177!5m2!1sen!2ssg",
    },
    {
        "id": "games-haven-amk-map",
        "name": "Games Haven - Ang Mo Kio",
        "address": "51 Ang Mo Kio Ave 3, #03-01, Singapore 569922",
        "iframe": "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d63821.04849133675!2d103.76901934863278!3d1.2846211999999948!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da17d78ce21027%3A0x2ce73b73e4a0a9b7!2sGames%20Haven%20-%20Ang%20Mo%20Kio!5e0!3m2!1sen!2ssg!4v1702821196724!5m2!1sen!2ssg",
    },
    {
        "id": "games-haven-je-map",
        "name": "Games Haven - Jurong East",
        "address": "131 Jurong Gateway Rd, #02-245, Singapore 600131",
        "iframe": "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d63818.86192364747!2d103.77101964863279!3d1.3693644999999974!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da10057f577829%3A0x63e8b2fbab922947!2sGames%20Haven%20-%20Jurong%20East!5e0!3m2!1sen!2ssg!4v1702821214358!5m2!1sen!2ssg",
    },
    {
        "id": "grey-ogre-map",
        "name": "Grey Ogre Games",
        "address": "83 Club St, Singapore 069451",
        "iframe": "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.8199964760065!2d103.84085797576442!3d1.2817574584814586!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da190d242b70db%3A0x965b932c3bc19eda!2sGrey%20Ogre%20Games!5e0!3m2!1sen!2ssg!4v1702821297360!5m2!1sen!2ssg",
    },
    {
        "id": "hideout-map",
        "name": "Hideout",
        "address": "803 King George's Ave, #02-190, Singapore 200803",
        "iframe": "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d15955.112777358516!2d103.84179288715819!3d1.3083185!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da19e075f4c4f5%3A0x60e4a2c61816be63!2sHideout!5e0!3m2!1sen!2ssg!4v1702821327690!5m2!1sen!2ssg",
    },
    {
        "id": "manapro-map",
        "name": "Mana Pro",
        "address": "BLK 203 Choa Chu Kang Ave 1, B1-41, Singapore 680203",
        "iframe": "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.6584888121897!2d103.74693327451605!3d1.3815577614740542!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da1176665e2737%3A0x3b8608ab4d67724f!2sMana%20Pro!5e0!3m2!1sen!2ssg!4v1702821359528!5m2!1sen!2ssg",
    },
    {
        "id": "mox-map",
        "name": "Mox & Lotus",
        "address": "215 Bedok North Street 1, #02-85, Singapore 460215",
        "iframe": "https://www.google.com/maps/embed?pb=!1m14!1m8!1m3!1d15954.999958678827!2d103.9334704!3d1.3259392!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da19d89d198d6b%3A0xb3e238feedd6c90d!2sMox%20%26%20Lotus!5e0!3m2!1sen!2ssg!4v1730797737444!5m2!1sen!2ssg",
    },
    {
        "id": "mtg-asia-map",
        "name": "MTG Asia",
        "address": "261 Waterloo St, #03-28, Singapore 180261",
        "iframe": "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.7930896678654!2d103.8493947744998!3d1.2989162986887468!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da19bb4a2bee83%3A0x28725aa3a3e2a51!2sMTG-Asia!5e0!3m2!1sen!2ssg!4v1703085334392!5m2!1sen!2ssg",
    },
    {
        "id": "onemtg-map",
        "name": "One MTG",
        "address": "100 Jln Sultan, #03-11 Sultan Plaza, Singapore 199001",
        "iframe": "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.7866900551694!2d103.85910407451628!3d1.3029641617257042!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da19180d91f3a1%3A0x75c807bf93d430a4!2sOne%20MTG!5e0!3m2!1sen!2ssg!4v1702821425238!5m2!1sen!2ssg",
    },
    {
        "id": "tefuda-map",
        "name": "Tefuda",
        "address": "B1-02 Macpherson Mall, Singapore 368125",
        "iframe": "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.740634996764!2d103.8765490749657!3d1.3317319986556433!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da178ea031acdb%3A0xac7ea94397d6a870!2sTefuda!5e0!3m2!1sen!2ssg!4v1743179304416!5m2!1sen!2ssg",
    }
]

// Map page
let mapListHtml = "";
let mapItemsHtml = "";
let topBtnHtml = `<a href="#map-list" class="btn btn-primary" role="button">Back to top</a>`;
let closeBtnHtml = `<button type="button" class="btn btn-secondary" style="margin-left: 7px;" data-bs-dismiss="modal">Close</button>`;

mapListHtml += `<div class="mb-4"><ul style="padding-left: 1rem">`;
for(let i = 0; i < lgsMap.length; i++) {
    mapListHtml += `<li><a href="#`+lgsMap[i].id+`" class="link-offset-2" alt="`+lgsMap[i].name+`">`+lgsMap[i].name+`</a></li>`;
    mapItemsHtml += `
        <div id="`+lgsMap[i].id+`" class="mb-4 map-item">
            <h5>`+lgsMap[i].name+`</h5>
            <div class="mb-2">`+lgsMap[i].address+`</div>
            <iframe class="w-100 h-100 border border-dark mb-3" src="`+lgsMap[i].iframe+`" allowfullscreen="" loading="lazy" referrerpolicy="no-referrer-when-downgrade"></iframe>
            <div>`+topBtnHtml+closeBtnHtml+`</div>
        </div>    
    `;
}
mapListHtml += `</ul></div>`;
document.getElementById("map").getElementsByClassName("modal-body")[0].innerHTML = mapListHtml + mapItemsHtml;

// Init
setVersionAndClearStorage();
onloadCart();
updateCartNo();
updateCartPage();

function setVersionAndClearStorage() {
    if(localStorage.getItem('version') !== null && localStorage.getItem('version') !== undefined && localStorage.getItem('version') !== "") {
        if (localStorage.getItem("version") !== version) {
            // clear storage from previous version if needed
        }
    }
    localStorage.setItem("version", version);
}

function onloadCart() {
    if(localStorage.getItem('cart') !== null && localStorage.getItem('cart') !== undefined && localStorage.getItem('cart') !== "") {
        cart = JSON.parse(localStorage.getItem('cart'));
    }
}

function removeFromCart(index) {
    // get from storage first in case multiple tabs add / removing
    if(localStorage.getItem('cart') !== null && localStorage.getItem('cart') !== undefined && localStorage.getItem('cart') !== "") {
        cart = JSON.parse(localStorage.getItem('cart'));
    } else {
        cart = [];
    }

    if (index >= 0 && cart.length > index) {
        cart.splice(index, 1);
        localStorage.setItem("cart", JSON.stringify(cart));
        updateCartPage();
    }
}

function updateCartPage() {
    let cartContent = document.getElementById("cartContent");
    cartContent.innerHTML = "";
    let html = "";
    if (cart.length > 0) {
        html += `<div class="row">`;

        for(let i=0; i<cart.length; i++) {
            let removeFromCartBtn = `<button data-index="`+i+`" type="button" class="removeFromCartBtn btn btn-danger btn-sm removeFromCartBtn"><i data-feather="trash-2" class="cartIcon"></i> Remove</button>`;
            let searchBtn = `<a href="/?s=`+encodeURIComponent(cart[i]["name"])+`&src=`+encodeURIComponent(cart[i]["src"])+`" class="btn btn-primary btn-sm cartSearchBtn ms-1"><i data-feather="search" class="cartIcon"></i> Search</a>`;

            html += `
            <div class="col-6 mb-3">
                <div class="text-center mb-2">
                    <a href="`+cart[i]["url"]+`" target="_blank">
                        <img src="`+(cart[i]["img"]===""?`https://placehold.co/304x424?text=`+cart[i]["name"]:cart[i]["img"])+`" loading="lazy" class="img-fluid w-100" alt="`+cart[i]["name"]+`"/>
                    </a>
                </div>
                <div class="text-center">
                    <div class="fs-6 lh-sm fw-bold mb-1">`+cart[i]["name"]+`</div>
                    `+((cart[i].hasOwnProperty("extraInfo") && cart[i]["extraInfo"]!=="")?`<div class="fs-6 lh-sm fw-bold mb-1">`+cart[i]["extraInfo"]+`</div>`:``)+`                    
                    `+((cart[i].hasOwnProperty("quality") && cart[i]["quality"]!=="")?`<div class="fs-6 lh-sm fw-bold mb-1">≪ `+cart[i]["quality"]+` ≫</div>`:``)+`
                    <div class="fs-6 lh-sm">S$ `+cart[i]["price"].toFixed(2)+`</div>
                    <div class="mb-2"><a href="`+cart[i]["url"]+`" target="_blank" class="link-offset-2">`+cart[i]["src"]+`</a></div>
                    <div>`+removeFromCartBtn+searchBtn+`</div>
                </div>
            </div>
            `;
        }
        html += `</div>`;

        if (cart.length >=2) {
            html += `<div class="mt-5"><button type="button" id="clearCartBtn" class="btn btn-danger w-100 text-uppercase">Remove all saved cards</button></div>`;
        }
    } else {
        html += `<strong>No cards saved yet.</strong>`;
    }
    cartContent.innerHTML = html;
    feather.replace();
    removeCartEventListeners();
    updateCartNo();
}

function updateCartNo() {
    let cartNo = document.getElementById("cartNo");

    if (cart.length > 0) {
        cartNo.innerHTML = ` (` + cart.length + `)`;
    } else {
        cartNo.innerHTML = ``;
    }
}

function removeCartEventListeners() {
    let removeFromCartBtns = document.querySelectorAll("button.removeFromCartBtn");
    removeFromCartBtns.forEach(function(elem) {
        elem.addEventListener("click", function() {
            if (this.getAttribute("data-index") !== "") {
                removeFromCart(this.getAttribute("data-index"));
                updateCartPage();
            }
        });
    });

    let emptyCartBtn = document.getElementById("clearCartBtn");
    if (emptyCartBtn !== null) {
        emptyCartBtn.addEventListener("click", function() {
            if(confirm("Are you sure you want to remove all saved cards?")) {
                if (cart.length > 0) {
                    cart = [];
                    localStorage.removeItem("cart");
                    updateCartPage();
                }
            }
        });
    }
}