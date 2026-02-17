import React, { useState, useEffect, useRef } from 'react';
import {
  Form,
  FloatingLabel,
  Offcanvas,
  Modal,
  Button
} from 'react-bootstrap';
import {
  FolderPlus,
  Map as MapIcon,
  HelpCircle,
  ArrowUp,
  Trash2,
  Search as SearchIcon,
  CheckSquare
} from 'react-feather';
import './index.css';

// --- Constants ---
const PAGE_TITLE = "Gishath Fetch: MTG Price Checker for Singapore's LGS";
const LGS_OPTIONS = [
  "5 Mana", "Agora Hobby", "Arcane Sanctum", "Card Affinity", "Cardboard Crack Games",
  "Cards Citadel", "Cards & Collections", "Dueller's Point", "Flagship Games",
  "Games Haven", "Grey Ogre Games", "Hideout", "Mana Pro", "Mox & Lotus",
  "MTG Asia", "OneMtg", "Tefuda", "The TCG Marketplace"
];

const LGS_MAP = [
  { id: "5-mana-map", name: "5 Mana", address: "511 Guillemard Rd, #02-06, Singapore 399849", iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.7686522542544!2d103.88875987494157!3d1.314306298673231!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da19ef83dc5edf%3A0xf45523d5c3efb509!2s5-MANA.SG!5e0!3m2!1sen!2ssg!4v1768142747318!5m2!1sen!2ssg", website: "https://5-mana.sg/" },
  { id: "agora-map", name: "Agora Hobby", address: "French Rd, #05-164 Blk 809, Singapore 200809", iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.778050505021!2d103.85967687451628!3d1.3084089617085968!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da19c9f7d7f74d%3A0xeaa1a66df7d4bcd6!2sAgora%20Hobby!5e0!3m2!1sen!2ssg!4v1702820213937!5m2!1sen!2ssg", website: "https://agorahobby.com/" },
  { id: "arcane-sanctum-map", name: "Arcane Sanctum", address: "809 French Rd, #02-36 Kitchener Complex, Singapore 200809", iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.778059032544!2d103.8596768749415!3d1.3084035986791807!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da197e8c761a49%3A0x8c56b7150064528b!2sArcane%20Sanctum!5e0!3m2!1sen!2ssg!4v1768317836907!5m2!1sen!2ssg", website: "https://arcanesanctumtcg.com/" },
  { id: "cardboard-crack-games-map", name: "Cardboard Crack Games", address: "Upper Bukit Timah Rd, #03-28 Beauty World Centre, 144, Singapore 588177", iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.7233430292667!2d103.7736843749657!3d1.3423740986449086!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da1920d676db93%3A0xe7b298b897da7b52!2sCardboard%20Crack%20Games!5e0!3m2!1sen!2ssg!4v1731824736033!5m2!1sen!2ssg", website: "https://www.cardboardcrackgames.com/" },
  { id: "cards-citadel-map", name: "Cards Citadel", address: "464 Crawford Ln, #02-01, Singapore 190464", iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.783678524258!2d103.85966947451631!3d1.3048646617197366!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da190c9e183751%3A0xa2119a95d1e683f2!2sCards%20Citadel!5e0!3m2!1sen!2ssg!4v1702820792196!5m2!1sen!2ssg", website: "https://cardscitadel.com/" },
  { id: "dueller-point-map", name: "Dueller's Point", address: "450 Hougang Ave 10, B1-541, Singapore 530450", iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.662159756766!2d103.89300967451602!3d1.3793695614811952!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da163eecb250ff%3A0xc7c259e72671dc62!2sDueller&#39;s%20Point!5e0!3m2!1sen!2ssg!4v1702820876967!5m2!1sen!2ssg", website: "https://www.duellerspoint.com/" },
  { id: "flagship-games-map", name: "Flagship Games", address: "214 Bishan St. 23, B1-223, Singapore 570214", iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d490.0218996351789!2d103.84829838647084!3d1.3574649065942905!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da173ef6ffcc0b%3A0x880386dee363a253!2sFlagship%20Games!5e0!3m2!1sen!2ssg!4v1734958555684!5m2!1sen!2ssg", website: "https://www.flagshipgames.sg/" },
  { id: "games-haven-pl-map", name: "Games Haven - Paya Lebar", address: "736 Geylang Rd, Singapore 389647", iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d63819.358332241325!2d103.79905633083244!3d1.350592080054757!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da1817d10ac901%3A0x2cacb3a0679089a2!2sGames%20Haven!5e0!3m2!1sen!2ssg!4v1702821045126!5m2!1sen!2ssg", website: "https://www.gameshaventcg.com/" },
  { id: "grey-ogre-map", name: "Grey Ogre Games", address: "83 Club St, Singapore 069451", iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.8199964760065!2d103.84085797576442!3d1.2817574584814586!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da190d242b70db%3A0x965b932c3bc19eda!2sGrey%20Ogre%20Games!5e0!3m2!1sen!2ssg!4v1702821297360!5m2!1sen!2ssg", website: "https://www.greyogregames.com/" },
  { id: "hideout-map", name: "Hideout", address: "803 King George's Ave, #02-190, Singapore 200803", iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d15955.112777358516!2d103.84179288715819!3d1.3083185!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da19e075f4c4f5%3A0x60e4a2c61816be63!2sHideout!5e0!3m2!1sen!2ssg!4v1702821327690!5m2!1sen!2ssg", website: "https://hideoutcg.com/" },
  { id: "manapro-map", name: "Mana Pro", address: "BLK 203 Choa Chu Kang Ave 1, B1-41, Singapore 680203", iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.6584888121897!2d103.74693327451605!3d1.3815577614740542!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da1176665e2737%3A0x3b8608ab4d67724f!2sMana%20Pro!5e0!3m2!1sen!2ssg!4v1702821359528!5m2!1sen!2ssg", website: "https://sg-manapro.com/" },
  { id: "mox-map", name: "Mox & Lotus", address: "215 Bedok North Street 1, #02-85, Singapore 460215", iframe: "https://www.google.com/maps/embed?pb=!1m14!1m8!1m3!1d15954.999958678827!2d103.9334704!3d1.3259392!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da19d89d198d6b%3A0xb3e238feedd6c90d!2sMox%20%26%20Lotus!5e0!3m2!1sen!2ssg!4v1730797737444!5m2!1sen!2ssg", website: "https://www.moxandlotus.sg/" },
  { id: "mtg-asia-map", name: "MTG Asia", address: "261 Waterloo St, #03-28, Singapore 180261", iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.7930896678654!2d103.8493947744998!3d1.2989162986887468!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da19bb4a2bee83%3A0x28725aa3a3e2a51!2sMTG-Asia!5e0!3m2!1sen!2ssg!4v1703085334392!5m2!1sen!2ssg", website: "https://www.mtg-asia.com/" },
  { id: "onemtg-map", name: "One MTG", address: "100 Jln Sultan, #03-11 Sultan Plaza, Singapore 199001", iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.7866900551694!2d103.85910407451628!3d1.3029641617257042!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da19180d91f3a1%3A0x75c807bf93d430a4!2sOne%20MTG!5e0!3m2!1sen!2ssg!4v1702821425238!5m2!1sen!2ssg", website: "https://onemtg.com.sg/" },
  { id: "tefuda-map", name: "Tefuda", address: "B1-02 Macpherson Mall, Singapore 368125", iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.740634996764!2d103.8765490749657!3d1.3317319986556433!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da178ea031acdb%3A0xac7ea94397d6a870!2sTefuda!5e0!3m2!1sen!2ssg!4v1743179304416!5m2!1sen!2ssg", website: "https://tefudagames.com/" },
  { id: "unsleeved-map", name: "Unsleeved", address: "17A Jln Klapa, #02-01, Singapore 199329", iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.7846024345927!2d103.85963729999999!3d1.3042818999999999!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da1963987b59bf%3A0xc1ed652c0bc65836!2sUnsalted%20by%20Lazy%20Potato!5e0!3m2!1sen!2ssg!4v1759481675787!5m2!1sen!2ssg", website: "https://hitpay.shop/unsleeved/" }
];

const API_BASE_URL = (window.location.hostname === "staging.gishathfetch.com" || window.location.hostname === "localhost")
  ? "https://staging-api.gishathfetch.com/"
  : "https://api.gishathfetch.com/";

const BASE_URL = (window.location.hostname === "staging.gishathfetch.com" || window.location.hostname === "localhost")
  ? "https://staging.gishathfetch.com/"
  : "https://gishathfetch.com/";

export default function App() {
  // --- State ---
  const [searchQuery, setSearchQuery] = useState("");
  const [suggestions, setSuggestions] = useState([]);
  const [selectedStores, setSelectedStores] = useState([]);
  const [searchResults, setSearchResults] = useState([]);
  const [isSearching, setIsSearching] = useState(false);
  const [searchProgress, setSearchProgress] = useState("");
  const [cart, setCart] = useState([]);

  // UI State
  const [showCart, setShowCart] = useState(false);
  const [showMap, setShowMap] = useState(false);
  const [showFaq, setShowFaq] = useState(false);
  const [showPrivacy, setShowPrivacy] = useState(false);
  const [showSuggestions, setShowSuggestions] = useState(false);

  const debounceTimer = useRef(null);

  // --- Initialization ---
  useEffect(() => {
    // 1. Initial Load from LocalStorage
    const storedLgs = localStorage.getItem('lgsSelected');
    if (storedLgs) {
      setSelectedStores(decodeURIComponent(storedLgs).split(","));
    } else {
      setSelectedStores(LGS_OPTIONS);
    }

    const storedCart = localStorage.getItem('cart');
    if (storedCart) {
      setCart(JSON.parse(storedCart));
    }

    // 2. Initial Search from URL
    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.has('s') && urlParams.get('s') !== "") {
      const q = decodeURIComponent(urlParams.get('s'));
      setSearchQuery(q);

      let stores = LGS_OPTIONS;
      if (urlParams.has('src') && LGS_OPTIONS.includes(decodeURIComponent(urlParams.get('src')))) {
        stores = [decodeURIComponent(urlParams.get('src'))];
        setSelectedStores(stores);
        localStorage.setItem("lgsSelected", encodeURIComponent(stores.join(",")));
      }

      // Delay search slightly to ensure state is set
      setTimeout(() => performSearch(q, stores), 100);
    }
  }, []);

  // --- Handlers ---
  const performSearch = (query, stores) => {
    if (!query || query.length < 3) return;

    setIsSearching(true);
    setSearchProgress("Searching LGS");
    setSearchResults([]);

    // Analytics
    if (window.gtag) {
      window.gtag('event', 'search', { 'search_term': query.toLowerCase() });
    }

    const searchUrl = `${API_BASE_URL}?s=${encodeURIComponent(query.toLowerCase())}&lgs=${encodeURIComponent(stores.join(','))}`;

    // Progress Animation Simulation
    let progressInterval = setInterval(() => {
      setSearchProgress(prev => prev.length > 25 ? "Searching LGS" : prev + " .");
    }, 1000);

    fetch(searchUrl)
      .then(res => res.json())
      .then(result => {
        if (result && result.data) {
          setSearchResults(result.data);
          updateUrlAndTitle(query);

          if (window.gtag) {
            window.gtag('event', 'view_search_results', { 'search_term': query.toLowerCase() });
          }
        }
      })
      .catch(err => console.error("Search error:", err))
      .finally(() => {
        setIsSearching(false);
        clearInterval(progressInterval);
      });
  };

  const updateUrlAndTitle = (query) => {
    if (window.location.hostname !== "localhost") {
      const newUrl = `${BASE_URL}?s=${encodeURIComponent(query.toLowerCase())}`;
      window.history.pushState(query.toLowerCase(), `${query.toLowerCase()} | ${PAGE_TITLE}`, newUrl);
      document.title = `${query.toLowerCase()} | ${PAGE_TITLE}`;
    }
  };

  const handleSearchSubmit = (e) => {
    e.preventDefault();
    performSearch(searchQuery, selectedStores);
  };

  const handleStoreToggle = (store) => {
    let newStores;
    if (selectedStores.includes(store)) {
      newStores = selectedStores.filter(s => s !== store);
    } else {
      newStores = [...selectedStores, store];
    }
    setSelectedStores(newStores);
    localStorage.setItem("lgsSelected", encodeURIComponent(newStores.join(",")));
  };

  const handleSelectAll = () => {
    setSelectedStores(LGS_OPTIONS);
    localStorage.setItem("lgsSelected", encodeURIComponent(LGS_OPTIONS.join(",")));
  };

  const handleSelectNone = () => {
    setSelectedStores([]);
    localStorage.setItem("lgsSelected", "");
  };

  const handleQueryChange = (e) => {
    const val = e.target.value;
    setSearchQuery(val);

    if (val.trim().length > 2) {
      clearTimeout(debounceTimer.current);
      debounceTimer.current = setTimeout(() => {
        fetch(`https://api.scryfall.com/cards/autocomplete?q=${encodeURIComponent(val.toLowerCase())}`)
          .then(res => res.json())
          .then(data => {
            if (data && data.data) {
              setSuggestions(data.data);
              setShowSuggestions(true);
            }
          });
      }, 300);
    } else {
      setSuggestions([]);
      setShowSuggestions(false);
    }
  };

  const handleSuggestionClick = (suggestion) => {
    setSearchQuery(suggestion);
    setShowSuggestions(false);
  };

  const addToCart = (card) => {
    const newCart = [card, ...cart];
    setCart(newCart);
    localStorage.setItem("cart", JSON.stringify(newCart));
  };

  const removeFromCart = (index) => {
    const newCart = [...cart];
    newCart.splice(index, 1);
    setCart(newCart);
    localStorage.setItem("cart", JSON.stringify(newCart));
  };

  const clearCart = () => {
    if (window.confirm("Are you sure you want to remove all saved cards?")) {
      setCart([]);
      localStorage.removeItem("cart");
    }
  };

  const isCardInCart = (card) => {
    return cart.some(item => JSON.stringify(item) === JSON.stringify(card));
  };

  // --- Render Helpers ---
  const renderCard = (card, index, isCart = false) => {
    const qualityFoil = [];
    if (card.quality) qualityFoil.push(`≪ ${card.quality} ≫`);
    if (card.isFoil) qualityFoil.push(<span key="foil" className='text-nowrap'>≪ <span className='animated-rainbow'>FOIL</span> ≫</span>);

    return (
      <div className={`col-6 col-lg-${isCart ? 6 : 3} mb-4`} key={index}>
        <div className="text-center mb-2">
          <a href={card.url} target="_blank" rel="noreferrer">
            <img
              src={card.img || `https://placehold.co/304x424?text=${encodeURIComponent(card.name)}`}
              loading="lazy"
              className="img-fluid w-100"
              alt={card.name}
            />
          </a>
        </div>
        <div className="text-center">
          <div className="fs-6 lh-sm fw-bold mb-1">{card.name}</div>
          {card.extraInfo && <div className="fs-6 lh-sm fw-bold mb-1">{card.extraInfo}</div>}
          {qualityFoil.length > 0 && (
            <div className="d-flex flex-wrap justify-content-center gap-1 fs-6 lh-sm fw-bold mb-1">
              {qualityFoil}
            </div>
          )}
          <div className="fs-6 lh-sm">S$ {card.price.toFixed(2)}</div>
          <div className="mb-2">
            <a href={card.url} target="_blank" rel="noreferrer" className="link-offset-2">
              {card.src}
            </a>
          </div>
          <div>
            {isCart ? (
              <div className="d-flex justify-content-center gap-1">
                <button
                  type="button"
                  className="removeFromCartBtn btn btn-danger btn-sm removeFromCartBtn"
                  onClick={() => removeFromCart(index)}
                >
                  <Trash2 size={12} className="cartIcon" /> Remove
                </button>
                <a
                  href={`${BASE_URL}?s=${encodeURIComponent(card.name)}&src=${encodeURIComponent(card.src)}`}
                  className="btn btn-primary btn-sm cartSearchBtn ms-1"
                  onClick={(e) => { e.preventDefault(); setShowCart(false); setSearchQuery(card.name); performSearch(card.name, [card.src]); }}
                >
                  <SearchIcon size={12} className="cartIcon" /> Search
                </a>
              </div>
            ) : (
              isCardInCart(card) ? (
                <button
                  type="button"
                  className="btn btn-success btn-sm addCartBtn"
                  disabled
                >
                  <CheckSquare size={12} className="cartIcon" /> Saved
                </button>
              ) : (
                <button
                  type="button"
                  className="addToCartBtn btn btn-primary btn-sm addCartBtn"
                  onClick={() => addToCart(card)}
                >
                  <FolderPlus size={12} className="cartIcon" /> Save
                </button>
              )
            )}
          </div>
        </div>
      </div>
    );
  };

  // --- Main Render ---
  return (
    <div id="top" className="container-xl my-3 px-3 pb-3">
      {/* Header */}
      <div className="mb-3 text-center">
        <div className="d-flex flex-row align-items-center justify-content-center mb-1">
          <div>
            <a href="/">
              <img id="logo" src="img/gishath-fetch-web.png" className="mb-2" alt="Gishath Fetch" />
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

      {/* Search Form */}
      <div>
        <form id="searchForm" onSubmit={handleSearchSubmit}>
          <div className="mb-3 position-relative">
            <div className="form-floating">
              <input
                type="search"
                className="form-control"
                id="search"
                placeholder="lightning bolt"
                value={searchQuery}
                onChange={handleQueryChange}
                autoComplete="off"
                onFocus={() => searchQuery.length > 2 && setShowSuggestions(true)}
              />
              <label htmlFor="search">Card Name</label>
            </div>

            {showSuggestions && suggestions.length > 0 && (
              <div id="suggestions" className="suggestions d-block">
                {suggestions.map((s, i) => (
                  <div key={i} className="suggestion-item" onClick={() => handleSuggestionClick(s)}>
                    {s}
                  </div>
                ))}
              </div>
            )}
          </div>

          <div><h6>Stores</h6></div>
          <div id="lgsCheckboxes">
            {LGS_OPTIONS.map((store, i) => (
              <div className="form-check form-check-inline" key={i}>
                <input
                  className="form-check-input lgsCheckbox"
                  type="checkbox"
                  id={`lgsCheckbox${i}`}
                  value={store}
                  checked={selectedStores.includes(store)}
                  onChange={() => handleStoreToggle(store)}
                />
                <label className="form-check-label" htmlFor={`lgsCheckbox${i}`}>
                  {store}
                </label>
              </div>
            ))}
          </div>

          <div className="mb-3">
            <a role="button" className="p-0 me-3 text-decoration-none" onClick={handleSelectAll}>All</a>
            <a role="button" className="p-0 text-decoration-none" onClick={handleSelectNone}>None</a>
          </div>

          <div className="mb-3 d-grid">
            <button
              id="searchBtn"
              type="submit"
              className="btn btn-primary"
              disabled={isSearching}
            >
              {isSearching ? searchProgress : "Search"}
            </button>
          </div>
        </form>
      </div>

      {/* Results */}
      {searchResults.length > 0 && (
        <div id="resultCount" className="mb-3 text-center bg-warning-subtle text-dark rounded py-2">
          {searchResults.length} result{searchResults.length > 1 ? "s" : ""} found
        </div>
      )}
      <div id="result" className="mb-3 text-center">
        <div className="row">
          {searchResults.map((card, i) => renderCard(card, i))}
        </div>
      </div>

      {/* Footer / Ads */}
      <div className="ad-large mt-4 pb-5 text-center d-print-none d-block d-sm-block">
        <div className="text-secondary mb-2" style={{ fontSize: '11px' }}>Advertisement</div>
        <div style={{ minHeight: '90px' }}>
          {/* AdSense slot placeholder */}
          <ins className="adsbygoogle" style={{ display: 'inline-block', width: '728px', height: '90px' }}
            data-ad-client="ca-pub-2393161407259792" data-ad-slot="6707964257"></ins>
        </div>
        <div className="text-center mt-2" style={{ fontSize: '11px' }}>
          <a href="https://www.patreon.com/GishathFetch" target="_blank" rel="noreferrer">Follow / Support Gishath Fetch on Patreon</a>
        </div>
      </div>

      {/* Fixed Bottom Navigation */}
      <div className="fixed-bottom bg-primary text-light text-center">
        <div className="d-flex flex-row align-items-center justify-content-center">
          <a
            role="button"
            className="py-1 link-light link-offset-2 link-underline-opacity-0"
            onClick={() => setShowCart(true)}
          >
            <div className="px-3 py-1">
              <FolderPlus size={14} className="me-1 mb-1" /> Saved {cart.length > 0 && `(${cart.length})`}
            </div>
          </a>
          <a
            role="button"
            className="py-1 link-light link-offset-2 link-underline-opacity-0"
            onClick={() => setShowMap(true)}
          >
            <div className="px-3 py-1">
              <MapIcon size={14} className="me-1" /> Map
            </div>
          </a>
          <a
            role="button"
            className="py-1 link-light link-offset-2 link-underline-opacity-0"
            onClick={() => setShowFaq(true)}
          >
            <div className="px-3 py-1">
              <HelpCircle size={14} className="me-1 mb-1" /> FAQs
            </div>
          </a>
          <a href="#top" className="py-1 link-light link-offset-2 link-underline-opacity-0">
            <div className="px-3 py-1">
              <ArrowUp size={14} className="me-1" /> Top
            </div>
          </a>
        </div>
      </div>

      {/* --- Modals & Offcanvas --- */}

      {/* Saved Cards Offcanvas */}
      <Offcanvas show={showCart} onHide={() => setShowCart(false)} placement="end">
        <Offcanvas.Header closeButton>
          <Offcanvas.Title>Saved Cards</Offcanvas.Title>
        </Offcanvas.Header>
        <Offcanvas.Body>
          <div className="mb-3 small text-muted">
            When a card is saved, a snapshot of it from that point in time is taken. If there is any change in its price or availability, it will not be updated automatically.
          </div>
          {cart.length > 0 ? (
            <>
              <div className="row">
                {cart.map((card, i) => renderCard(card, i, true))}
              </div>
              {cart.length >= 2 && (
                <div className="mt-5">
                  <Button variant="danger" className="w-100 text-uppercase" onClick={clearCart}>
                    Remove all saved cards
                  </Button>
                </div>
              )}
            </>
          ) : (
            <strong>No cards saved yet.</strong>
          )}
        </Offcanvas.Body>
      </Offcanvas>

      {/* Map Modal */}
      <Modal show={showMap} onHide={() => setShowMap(false)} size="xl">
        <Modal.Header closeButton>
          <Modal.Title id="map-list">Where are the shops?</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="mb-4">
            <ul style={{ paddingLeft: '1rem' }}>
              {LGS_MAP.map((shop, i) => (
                <li key={i}><a href={`#${shop.id}`} className="link-offset-2">{shop.name}</a></li>
              ))}
            </ul>
          </div>
          {LGS_MAP.map((shop, i) => (
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
                <Button variant="secondary" className="ms-2" onClick={() => setShowMap(false)}>Close</Button>
              </div>
            </div>
          ))}
        </Modal.Body>
        <Modal.Footer className="justify-content-start">
          © 2023 gishathfetch.com by <a href="https://github.com/xenodus" target="_blank" rel="noreferrer">xenodus</a> | <Button variant="link" className="p-0" onClick={() => { setShowMap(false); setShowPrivacy(true); }}>privacy policy</Button>
        </Modal.Footer>
      </Modal>

      {/* FAQ Modal */}
      <Modal show={showFaq} onHide={() => setShowFaq(false)} size="xl">
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
          © 2023 gishathfetch.com by <a href="https://github.com/xenodus" target="_blank" rel="noreferrer">xenodus</a> | <Button variant="link" className="p-0" onClick={() => { setShowFaq(false); setShowPrivacy(true); }}>privacy policy</Button>
        </Modal.Footer>
      </Modal>

      {/* Privacy Modal */}
      <Modal show={showPrivacy} onHide={() => setShowPrivacy(false)} size="xl">
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

    </div>
  );
}
