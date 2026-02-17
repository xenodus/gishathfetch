import React from 'react';

const Header = () => {
    return (
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
    );
};

export default Header;
