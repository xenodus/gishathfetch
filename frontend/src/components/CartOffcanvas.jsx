import { useMemo, useState } from "react";
import { Button, Form, Modal, Offcanvas } from "react-bootstrap";
import { decodeCartImport, encodeCartExport } from "../utils/cartTransfer";
import AdComponent from "./AdComponent";
import Card from "./Card";

const CartOffcanvas = ({
  show,
  onHide,
  cart,
  isCardInCart,
  removeFromCart,
  onSearchStore,
  onSearchWithFavouriteStores,
  hasFavourites,
  onClearCart,
  onImportCart,
  baseUrl,
}) => {
  const [sortOption, setSortOption] = useState("default");
  const [showExportModal, setShowExportModal] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [exportCode, setExportCode] = useState("");
  const [importCode, setImportCode] = useState("");
  const [importError, setImportError] = useState(null);
  const [copyFeedback, setCopyFeedback] = useState(null);

  const handleClearCart = () => {
    if (window.confirm("Are you sure you want to remove all saved cards?")) {
      onClearCart();
    }
  };

  const handleOpenExport = () => {
    setExportCode(encodeCartExport(cart));
    setCopyFeedback(null);
    setShowExportModal(true);
  };

  const handleCopyExport = async () => {
    try {
      await navigator.clipboard.writeText(exportCode);
      setCopyFeedback("Copied to clipboard");
    } catch {
      setCopyFeedback("Select the code and copy manually");
    }
  };

  const handleOpenImport = () => {
    setImportCode("");
    setImportError(null);
    setShowImportModal(true);
  };

  const handleImport = () => {
    const { items, error } = decodeCartImport(importCode);
    if (error) {
      setImportError(error);
      return;
    }
    onImportCart(items);
    setShowImportModal(false);
    setImportCode("");
    setImportError(null);
  };

  const displayedCart = useMemo(() => {
    const withIndex = cart.map((card, index) => ({
      ...card,
      originalIndex: index,
    }));

    if (sortOption === "name-asc") {
      return [...withIndex].sort((a, b) =>
        (a.name || "").localeCompare(b.name || "", undefined, {
          sensitivity: "base",
        }),
      );
    }

    if (sortOption === "name-desc") {
      return [...withIndex].sort((a, b) =>
        (b.name || "").localeCompare(a.name || "", undefined, {
          sensitivity: "base",
        }),
      );
    }

    return withIndex;
  }, [cart, sortOption]);

  const groupedCart = useMemo(() => {
    if (sortOption !== "store") return null;

    const groups = {};
    cart.forEach((card, index) => {
      const storeName = card.src || "Unknown Store";
      if (!groups[storeName]) {
        groups[storeName] = { cards: [], total: 0 };
      }
      groups[storeName].cards.push({ ...card, originalIndex: index });
      groups[storeName].total += card.price || 0;
    });
    return groups;
  }, [cart, sortOption]);

  return (
    <Offcanvas show={show} onHide={onHide} placement="end">
      <Offcanvas.Header closeButton>
        <Offcanvas.Title>Saved Cards</Offcanvas.Title>
      </Offcanvas.Header>
      <Offcanvas.Body>
        <div className="mb-3 small text-muted">
          When a card is saved, a snapshot of it from that point in time is
          taken. If there is any change in its price or availability, it will
          not be updated automatically.
        </div>

        <div className="mb-4 d-flex flex-column gap-2">
          <div className="d-flex flex-column flex-sm-row gap-2">
            <Button
              variant="outline-primary"
              size="sm"
              className="flex-fill"
              onClick={handleOpenExport}
              disabled={cart.length === 0}
            >
              Export saved list
            </Button>
            <Button
              variant="outline-secondary"
              size="sm"
              className="flex-fill"
              onClick={handleOpenImport}
            >
              Import saved list
            </Button>
          </div>
          {cart.length > 0 && (
            <Button
              variant="outline-danger"
              size="sm"
              className="w-100 text-uppercase"
              onClick={handleClearCart}
            >
              Remove all saved cards
            </Button>
          )}
        </div>

        {cart.length > 0 && (
          <Form.Group className="mb-4">
            <Form.Label className="small fw-bold text-uppercase mb-1">
              Sort By
            </Form.Label>
            <Form.Select
              value={sortOption}
              onChange={(e) => setSortOption(e.target.value)}
              size="sm"
            >
              <option value="default">Saved Order</option>
              <option value="name-asc">Card Name Asc</option>
              <option value="name-desc">Card Name Desc</option>
              <option value="store">Store</option>
            </Form.Select>
          </Form.Group>
        )}

        {cart.length > 0 ? (
          <>
            {(sortOption === "default" ||
              sortOption === "name-asc" ||
              sortOption === "name-desc") && (
              <div className="row">
                {displayedCart.map((card) => (
                  <Card
                    key={card.originalIndex}
                    card={card}
                    index={card.originalIndex}
                    isCart={true}
                    isCardInCart={isCardInCart}
                    removeFromCart={removeFromCart}
                    onSearchStore={onSearchStore}
                    onSearchWithFavouriteStores={onSearchWithFavouriteStores}
                    hasFavourites={hasFavourites}
                    baseUrl={baseUrl}
                  />
                ))}
              </div>
            )}

            {sortOption === "store" &&
              Object.entries(groupedCart).map(([storeName, data]) => (
                <div key={storeName} className="mb-4">
                  <h5 className="border-bottom pb-2 mb-3">
                    {storeName} - S$ {data.total.toFixed(2)}
                  </h5>
                  <div className="row">
                    {data.cards.map((card) => (
                      <Card
                        key={card.originalIndex}
                        card={card}
                        index={card.originalIndex}
                        isCart={true}
                        isCardInCart={isCardInCart}
                        removeFromCart={removeFromCart}
                        onSearchStore={onSearchStore}
                        onSearchWithFavouriteStores={
                          onSearchWithFavouriteStores
                        }
                        hasFavourites={hasFavourites}
                        baseUrl={baseUrl}
                      />
                    ))}
                  </div>
                </div>
              ))}

            <div className="mt-4">
              <AdComponent lazyLoad />
            </div>
          </>
        ) : (
          <strong>No cards saved yet.</strong>
        )}
      </Offcanvas.Body>

      <Modal
        show={showExportModal}
        onHide={() => setShowExportModal(false)}
        centered
      >
        <Modal.Header closeButton>
          <Modal.Title>Export saved list</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p className="small text-muted">
            Copy this code and paste it on another device using Import saved
            list. It includes card snapshots stored in your browser only.
          </p>
          <Form.Control
            as="textarea"
            rows={6}
            readOnly
            value={exportCode}
            className="font-monospace small"
            onFocus={(event) => event.target.select()}
          />
          {copyFeedback && (
            <div className="small text-success mt-2">{copyFeedback}</div>
          )}
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowExportModal(false)}>
            Close
          </Button>
          <Button variant="primary" onClick={handleCopyExport}>
            Copy code
          </Button>
        </Modal.Footer>
      </Modal>

      <Modal
        show={showImportModal}
        onHide={() => setShowImportModal(false)}
        centered
      >
        <Modal.Header closeButton>
          <Modal.Title>Import saved list</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p className="small text-muted">
            Paste an export code from another device. Imported cards are merged
            into your current saved list; duplicates keep the newer snapshot.
          </p>
          <Form.Control
            as="textarea"
            rows={6}
            value={importCode}
            onChange={(event) => {
              setImportCode(event.target.value);
              if (importError) {
                setImportError(null);
              }
            }}
            className="font-monospace small"
            placeholder="Paste export code here"
          />
          {importError && (
            <div className="small text-danger mt-2">{importError}</div>
          )}
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowImportModal(false)}>
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={handleImport}
            disabled={!importCode.trim()}
          >
            Import and append
          </Button>
        </Modal.Footer>
      </Modal>
    </Offcanvas>
  );
};

export default CartOffcanvas;
