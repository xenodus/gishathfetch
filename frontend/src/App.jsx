import React, { useState, useEffect, useRef } from 'react';
import './index.css';

// --- Modular Components ---
import Header from './components/Header';
import SearchForm from './components/SearchForm';
import SearchResults from './components/SearchResults';
import CartOffcanvas from './components/CartOffcanvas';
import Modals from './components/Modals';
import Footer from './components/Footer';

// --- Constants ---
import {
  LGS_OPTIONS,
  LGS_MAP,
  BASE_URL
} from './constants';

// --- Hooks ---
import useCart from './hooks/useCart';
import useSearch from './hooks/useSearch';

export default function App() {
  const {
    cart,
    showCart,
    setShowCart,
    addToCart,
    removeFromCart,
    clearCart,
    isCardInCart
  } = useCart();

  const {
    searchQuery,
    setSearchQuery,
    isSearching,
    searchResults,
    searchProgress,
    suggestions,
    showSuggestions,
    setShowSuggestions,
    selectedStores,
    handleQueryChange,
    handleSuggestionClick,
    handleSearchSubmit,
    toggleStore,
    selectAllStores,
    selectNoStores,
    performSearch
  } = useSearch();

  const [modalType, setModalType] = useState(null);

  // --- Handlers ---
  const handleCardSearch = (cardName, sourceStore) => {
    setSearchQuery(cardName);
    setShowCart(false);
    performSearch(cardName, [sourceStore]);
  };

  // --- Main Render ---
  return (
    <div id="top" className="container-xl my-3 px-3 pb-3">
      <Header />

      <SearchForm
        searchQuery={searchQuery}
        onQueryChange={handleQueryChange}
        onSearchSubmit={handleSearchSubmit}
        suggestions={suggestions}
        showSuggestions={showSuggestions}
        onSuggestionClick={handleSuggestionClick}
        onFocus={() => searchQuery.length > 2 && setShowSuggestions(true)}
        isSearching={isSearching}
        searchProgress={searchProgress}
        lgsOptions={LGS_OPTIONS}
        selectedStores={selectedStores}
        onStoreToggle={toggleStore}
        onSelectAll={selectAllStores}
        onSelectNone={selectNoStores}
      />

      <SearchResults
        results={searchResults}
        isCardInCart={isCardInCart}
        addToCart={addToCart}
        removeFromCart={removeFromCart}
        onSearchStore={handleCardSearch}
        baseUrl={BASE_URL}
      />

      <Footer
        cartCount={cart.length}
        onShowCart={() => setShowCart(true)}
        onShowMap={() => setModalType('MAP')}
        onShowFaq={() => setModalType('FAQ')}
      />

      <CartOffcanvas
        show={showCart}
        onHide={() => setShowCart(false)}
        cart={cart}
        isCardInCart={isCardInCart}
        removeFromCart={removeFromCart}
        onSearchStore={handleCardSearch}
        onClearCart={clearCart}
        baseUrl={BASE_URL}
      />

      <Modals
        showMap={modalType === 'MAP'}
        onHideMap={() => setModalType(null)}
        showFaq={modalType === 'FAQ'}
        onHideFaq={() => setModalType(null)}
        showPrivacy={modalType === 'PRIVACY'}
        onHidePrivacy={() => setModalType(null)}
        onShowPrivacy={() => setModalType('PRIVACY')}
        lgsMapData={LGS_MAP}
      />
    </div>
  );
}
