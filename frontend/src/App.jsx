import React, { useState, lazy, Suspense, useCallback } from 'react';
import './index.css';

// --- Modular Components ---
import Header from './components/Header';
import SearchForm from './components/SearchForm';
import SearchResults from './components/SearchResults';
import Footer from './components/Footer';

const CartOffcanvas = lazy(() => import('./components/CartOffcanvas'));
const Modals = lazy(() => import('./components/Modals'));

// --- Constants ---
import {
  LGS_OPTIONS,
  LGS_MAP,
  BASE_URL,
  MIN_SEARCH_LENGTH
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
    hasSearched,
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
  const handleCardSearch = useCallback((e, cardName, sourceStore) => {
    if (e && e.preventDefault) e.preventDefault();
    setSearchQuery(cardName);
    setShowCart(false);
    setShowSuggestions(false); // Close suggestions dropdown
    performSearch(cardName, [sourceStore]);
  }, [performSearch, setSearchQuery, setShowCart, setShowSuggestions]);

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
        onFocus={() => searchQuery.length > MIN_SEARCH_LENGTH - 1 && setShowSuggestions(true)}
        isSearching={isSearching}
        searchProgress={searchProgress}
        lgsOptions={LGS_OPTIONS}
        selectedStores={selectedStores}
        onStoreToggle={toggleStore}
        onSelectAll={selectAllStores}
        onSelectNone={selectNoStores}
        onCloseSuggestions={() => setShowSuggestions(false)}
      />

      <SearchResults
        results={searchResults}
        isSearching={isSearching}
        hasSearched={hasSearched}
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

      <Suspense fallback={null}>
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
      </Suspense>
    </div>
  );
}
