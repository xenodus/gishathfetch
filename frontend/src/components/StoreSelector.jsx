import React, { memo } from 'react';

const StoreSelector = memo(({ options, selectedStores, onToggle, onSelectAll, onSelectNone }) => {
    return (
        <>
            <div><h6>Stores</h6></div>
            <div id="lgsCheckboxes">
                {options.map((store, i) => (
                    <div className="form-check form-check-inline" key={i}>
                        <input
                            className="form-check-input lgsCheckbox"
                            type="checkbox"
                            id={`lgsCheckbox${i}`}
                            value={store}
                            checked={selectedStores.includes(store)}
                            onChange={() => onToggle(store)}
                        />
                        <label className="form-check-label" htmlFor={`lgsCheckbox${i}`}>
                            {store}
                        </label>
                    </div>
                ))}
            </div>

            <div className="mb-3">
                <button type="button" className="btn btn-link p-0 me-3 text-decoration-none" onClick={onSelectAll}>All</button>
                <button type="button" className="btn btn-link p-0 text-decoration-none" onClick={onSelectNone}>None</button>
            </div>
        </>
    );
});

export default StoreSelector;
