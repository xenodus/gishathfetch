const pageTitle = "Gishath Fetch: MTG Price Checker for Singapore's LGS";
const form = document.getElementById("searchForm");
const lgsCheckboxesDiv = document.getElementById("lgsCheckboxes");
const searchInput = document.getElementById("search");
const suggestionsDiv = document.getElementById("suggestions");
const submitBtn = document.getElementById("submitBtn");
const resultDiv = document.getElementById("result");
const resultCountDiv = document.getElementById("resultCount");
const lgsCheckboxes = document.getElementsByName('lgs[]');
const lgsOptions = [
    "Agora Hobby",
    "Card Affinity",
    "Cardboard Crack Games",
    "Cards Citadel",
    "Cards & Collections",
    "Dueller's Point",
    "Flagship Games",
    "Games Haven",
    "Grey Ogre Games",
    "Hideout",
    "Mana Pro",
    "Mox & Lotus",
    "MTG Asia",
    "OneMtg",
    "The TCG Marketplace"
];
const alreadyInCartBtnHtml = `<i data-feather="check-square" class="cartIcon"></i> Saved`;

let contentAd = "";
let timeouts = [];
let searchResults = [];
let baseUrl = "https://gishathfetch.com/";
let apiBaseUrl = "https://api.gishathfetch.com/";

if (window.location.hostname === "staging.gishathfetch.com" || window.location.hostname === "localhost") {
    baseUrl = "https://staging.gishathfetch.com/";
    apiBaseUrl = "https://staging-api.gishathfetch.com/";
    document.head.innerHTML += `<meta name="robots" content="noindex" />`;
}

setupConfig();

// Pre-select checkboxes and pre-fill search from cookie
function setupConfig() {
    appendLgsCheckboxes();
    setupEventListeners();
    onloadSearch();
}

function onloadSearch() {
    const urlParams = new URLSearchParams(window.location.search);

    if (urlParams.has('s') && !urlParams.has('s', '')) {
        if (urlParams.has('src') && !urlParams.has('src', '') && lgsOptions.includes(decodeURIComponent(urlParams.get('src')))) {
            localStorage.setItem("lgsSelected", encodeURIComponent(urlParams.get('src')));
            appendLgsCheckboxes();
        }
        searchInput.value = decodeURIComponent(urlParams.get('s'));
        submitBtn.click();
    }
}

function setupEventListeners() {
    form.addEventListener("submit", searchCard);

    document.addEventListener("keypress", function(event) {
        if (event.keyCode === 13) {
            event.preventDefault();
            submitBtn.click();
        }
    });
}

function appendLgsCheckboxes() {
    let lgsSelected = [];

    if(localStorage.getItem('lgsSelected') !== null && localStorage.getItem('lgsSelected') !== undefined && localStorage.getItem('lgsSelected') !== "") {
        lgsSelected = decodeURIComponent(localStorage.getItem('lgsSelected')).split(",");
    } else {
        lgsSelected = lgsOptions;
    }

    lgsCheckboxesDiv.innerHTML = '';
    for(let i=0; i<lgsOptions.length; i++) {
        let isChecked = lgsSelected.includes(lgsOptions[i]) ? "checked" : "";
        lgsCheckboxesDiv.innerHTML += `
                <div class="form-check form-check-inline">
                  <input class="form-check-input" type="checkbox" id="lgsCheckbox`+i+`" class="lgsCheckboxes" value="`+lgsOptions[i]+`" name="lgs[]" `+isChecked+`>
                  <label class="form-check-label" for="lgsCheckbox`+i+`">`+lgsOptions[i]+`</label>
                </div>
              `;
    }
}

function clearTimeouts() {
    for (let i=0; i<timeouts.length; i++) {
        clearTimeout(timeouts[i]);
    }
}

// Timeout 15s in backend
function updateSubmitBtnProgress() {
    submitBtn.innerHTML = "Searching LGS"

    for(let i=1; i<=15; i++){
        timeouts.push(window.setTimeout(function(){
            submitBtn.innerHTML += " ."
        }, i * 1000));
    }
}

function resetResult() {
    // Copy ad for placement in content
    if (document.getElementsByClassName("ad-large").length > 0) {
        contentAd = document.getElementsByClassName("ad-large")[0].outerHTML;
    }

    resultDiv.innerHTML = "";
    resultCountDiv.innerHTML = "";
}

function resetSubmitBtn() {
    clearTimeouts();
    submitBtn.innerHTML = "Search";
    submitBtn.disabled = false;
}

function updatePageUrlTitle(searchStr, url) {
    if (window.location.hostname !== "localhost") {
        window.history.pushState(searchStr.toLowerCase(), searchStr.toLowerCase() + " | " + pageTitle, url);
        document.title = searchStr.toLowerCase() + " | " + pageTitle;
    }
}

function searchCard(event) {
    event.preventDefault();

    let searchStr = searchInput.value.trim()

    // End if empty search str
    if (searchStr === "" || searchStr.length < 3) {
        return
    }

    // Tag search str
    gtag('event', 'search', {
        'search_term': searchStr.toLowerCase()
    });

    let lgsSelected = [];

    for(let i=0; i<lgsCheckboxes.length; i++) {
        if (lgsCheckboxes[i].checked) {
            lgsSelected.push(lgsCheckboxes[i].value)
        }
    }

    if (lgsSelected.length === 0) {
        lgsSelected = lgsOptions;
        for(let i=0; i<lgsCheckboxes.length; i++) {
            lgsCheckboxes[i].checked = true;
        }
    }

    // Set state to disabled
    submitBtn.disabled = true;
    // Reset result div
    resetResult();

    let request = new XMLHttpRequest();
    let searchQueryString = "?s="+encodeURIComponent(searchStr.toLowerCase());
    let searchUrl = apiBaseUrl + searchQueryString
    searchUrl += "&lgs=" + encodeURIComponent(lgsSelected.join(','));

    localStorage.setItem("lgsSelected", encodeURIComponent(lgsSelected.join(",")));

    request.open("GET", searchUrl);
    request.send();

    updateSubmitBtnProgress();

    request.onreadystatechange = function() {
        if (request.readyState === XMLHttpRequest.DONE) {
            let resultCount = 0;

            // Check the status of the response
            if (request.status === 200) {
                // Access the data returned by the server
                let result = JSON.parse(request.responseText);
                // Do something with the data
                if (result.hasOwnProperty("data")) {
                    if (result["data"] !== null && result["data"].length > 0) {
                        searchResults = result["data"];
                        updatePageUrlTitle(searchStr, baseUrl + searchQueryString);
                        let html = `<div class="row">`;
                        for(let i = 0; i < result["data"].length; i++) {
                            if (result["data"][i].hasOwnProperty("url")
                                && result["data"][i].hasOwnProperty("img")
                                && result["data"][i].hasOwnProperty("name")
                                && result["data"][i].hasOwnProperty("price")
                                && result["data"][i].hasOwnProperty("src")) {

                                // add to cart btn state
                                let addToCartBtn = `<button data-index="`+i+`" type="button" class="addToCartBtn btn btn-primary btn-sm addCartBtn"><i data-feather="folder-plus" class="cartIcon"></i> Save</button>`;
                                if (existsInCart(result["data"][i]) === true) {
                                    addToCartBtn = `<button type="button" class="btn btn-success btn-sm addCartBtn" disabled>`+alreadyInCartBtnHtml+` </button>`;
                                }

                                let h = `
                                  <div class="col-lg-3 col-6 mb-4">
                                    <div class="text-center mb-2">
                                      <a href="`+result["data"][i]["url"]+`" target="_blank">
                                        <img src="`+(result["data"][i]["img"]===""?`https://placehold.co/304x424?text=`+result["data"][i]["name"]:result["data"][i]["img"])+`" loading="lazy" class="img-fluid w-100" alt="`+result["data"][i]["name"]+`"/>
                                      </a>
                                    </div>
                                    <div class="text-center">
                                      <div class="fs-6 lh-sm fw-bold mb-1">`+result["data"][i]["name"]+`</div>
                                      `+((result["data"][i].hasOwnProperty("extraInfo") && result["data"][i]["extraInfo"]!=="")?`<div class="fs-6 lh-sm fw-bold mb-1">`+result["data"][i]["extraInfo"]+`</div>`:``)+`
                                      `+((result["data"][i].hasOwnProperty("quality") && result["data"][i]["quality"]!=="")?`<div class="fs-6 lh-sm fw-bold mb-1">≪ `+result["data"][i]["quality"]+` ≫</div>`:``)+`
                                      <div class="fs-6 lh-sm">S$ `+result["data"][i]["price"].toFixed(2)+`</div>
                                      <div class="mb-2"><a href="`+result["data"][i]["url"]+`" target="_blank" class="link-offset-2">`+result["data"][i]["src"]+`</a></div>
                                      <div>`+addToCartBtn+`</div>
                                    </div>
                                  </div>`;

                                // Only place in content if result count > 8
                                if (result["data"].length > 8 && (((i+1)%8) === 0) && (i+1 !== result["data"].length)) {
                                    h += contentAd;
                                }

                                html += h;
                                resultCount++;
                            }
                        }
                        html += `</div>`;
                        resultDiv.innerHTML = html;
                    }
                    feather.replace();
                    addCartEventListeners();
                }

                // Tag search str
                gtag('event', 'view_search_results', {
                    'search_term': searchStr.toLowerCase()
                });

            } else {
                // Handle error
            }

            resultCountDiv.innerHTML = `<div class="py-2">`+resultCount+` result`+(resultCount>1?"s":"")+` found</div>`;

            // Reset state
            resetSubmitBtn();
        }
    };
}

function addCartEventListeners() {
    let addToCartBtns = document.querySelectorAll("button.addToCartBtn");
    addToCartBtns.forEach(function(elem) {
        elem.addEventListener("click", function() {
            if (this.getAttribute("data-index") !== "") {
                addToCart(this.getAttribute("data-index"));
                this.innerHTML = alreadyInCartBtnHtml;
                this.disabled = true;
                this.classList.remove("btn-primary");
                this.classList.add("btn-success");
                feather.replace();
            }
        });
    });
}

function addToCart(index) {
    if (index >= 0 && searchResults.length > index) {
        // get from storage first in case multiple tabs add / removing
        if(localStorage.getItem('cart') !== null && localStorage.getItem('cart') !== undefined && localStorage.getItem('cart') !== "") {
            cart = JSON.parse(localStorage.getItem('cart'));
        } else {
            cart = [];
        }

        cart.push(searchResults[index]);
        localStorage.setItem("cart", JSON.stringify(cart));
        updateCartPage();
    }
}

function existsInCart(item) {
    if (cart.length > 0) {
        for(let i=0; i<cart.length; i++) {
            if (JSON.stringify(cart[i]) === JSON.stringify(item)) {
                return true;
            }
        }
    }
    return false;
}

const debounceTimeout = 300;
let debounceTimer;

searchInput.addEventListener('input', () => {
    let searchStr = searchInput.value.trim();

    if (searchStr.length > 2) {
        clearTimeout(debounceTimeout);

        debounceTimer = setTimeout(() => {
            const request = new XMLHttpRequest();
            request.open('GET', `https://api.scryfall.com/cards/autocomplete?q=${encodeURIComponent(searchStr.toLowerCase())}`, true);
            request.onload = function () {
                if (request.status === 200) {
                    let result = JSON.parse(request.responseText);
                    if (result.hasOwnProperty("data")) {
                        displaySuggestions(boldMatchingSuggestions(result["data"], searchStr));
                    }
                } else {
                    console.error('There was an error making the request:', xhr.statusText);
                }
            };
            request.onerror = function () {
                console.error('Request failed');
            };
            request.send();
        }, debounceTimeout);
    } else {
        clearSuggestions()
    }
});

function displaySuggestions(suggestions) {
    clearSuggestions();
    suggestionsDiv.style.display = 'block';
    suggestions.forEach(suggestion => {
        const suggestionItem = document.createElement('div');
        suggestionItem.className = 'suggestion-item';
        suggestionItem.innerHTML = suggestion;
        suggestionItem.addEventListener('click', () => {
            searchInput.value = suggestionItem.innerText;
            clearSuggestions();
        });
        suggestionsDiv.appendChild(suggestionItem);
    });
}

function clearSuggestions() {
    suggestionsDiv.innerHTML = '';
    suggestionsDiv.style.display = 'none';
}

function boldMatchingSuggestions(suggestions, searchStr) {
    let boldedSuggestions = [];
    suggestions.forEach(suggestion => {
        let boldedSuggestion = suggestion.replace(new RegExp(searchStr, 'gi'), (match) => `<b>${match}</b>`);
        boldedSuggestions.push(boldedSuggestion);
    });
    return boldedSuggestions;
}

// Hide suggestions box when clicking outside
document.addEventListener('click', (event) => {
    if (!searchInput.contains(event.target) && !suggestionsDiv.contains(event.target)) {
        clearSuggestions();
    }
});
