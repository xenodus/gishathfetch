import React, { memo } from 'react';
import { FolderPlus, Search as SearchIcon, Trash2, CheckSquare } from 'react-feather';

const Card = memo(({
    card,
    index,
    isCart = false,
    isCardInCart,
    addToCart,
    removeFromCart,
    onSearchStore,
    baseUrl
}) => {
    const qualityFoil = [];
    if (card.quality) qualityFoil.push(`≪ ${card.quality} ≫`);
    if (card.isFoil) qualityFoil.push(<span key="foil" className='text-nowrap'>≪ <span className='animated-rainbow'>FOIL</span> ≫</span>);

    const inCart = isCardInCart(card);

    return (
        <div className={`col-6 col-lg-${isCart ? 6 : 3} mb-4`}>
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
                                className="removeFromCartBtn btn btn-danger btn-sm"
                                onClick={() => removeFromCart(index)}
                            >
                                <Trash2 size={12} className="cartIcon" /> Remove
                            </button>
                            <a
                                href={`${baseUrl}?s=${encodeURIComponent(card.name)}&src=${encodeURIComponent(card.src)}`}
                                className="btn btn-primary btn-sm cartSearchBtn ms-1"
                                onClick={(e) => onSearchStore(e, card.name, card.src)}
                            >
                                <SearchIcon size={12} className="cartIcon" /> Search
                            </a>
                        </div>
                    ) : (
                        inCart ? (
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
});

export default Card;
