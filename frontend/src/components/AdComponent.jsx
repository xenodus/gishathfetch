import React, { useEffect } from 'react';

const AdComponent = () => {
    useEffect(() => {
        try {
            (window.adsbygoogle = window.adsbygoogle || []).push({});

            // Parity fix: use setTimeout to set z-index, exactly as in legacy index.js
            setTimeout(() => {
                const ads = document.querySelectorAll('ins.adsbygoogle');
                ads.forEach(ad => {
                    ad.style.zIndex = '1000';
                });
            }, 1000);
        } catch (e) {
            console.error("AdSense error:", e);
        }
    }, []);

    return (
        <div className="ad-large mt-4 pb-5 text-center d-print-none d-block d-sm-block w-100">
            <div className="text-secondary mb-2" style={{ fontSize: '11px' }}>Advertisement</div>
            <div style={{ minHeight: '90px' }}>
                <ins className="adsbygoogle"
                    style={{ display: 'inline-block', width: '728px', height: '90px' }}
                    data-ad-client="ca-pub-2393161407259792"
                    data-ad-slot="6707964257"></ins>
            </div>
            <div className="text-center mt-2" style={{ fontSize: '11px' }}>
                <a href="https://www.patreon.com/GishathFetch" target="_blank" rel="noreferrer">Follow / Support Gishath Fetch on Patreon</a>
            </div>
        </div>
    );
};

export default AdComponent;
