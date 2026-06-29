import { useEffect, useRef, useState } from "react";

const LazyMapIframe = ({ src, title, isActive }) => {
  const containerRef = useRef(null);
  const [shouldLoad, setShouldLoad] = useState(false);

  useEffect(() => {
    if (!isActive || shouldLoad) {
      return;
    }

    const container = containerRef.current;
    if (!container) {
      return;
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setShouldLoad(true);
          observer.disconnect();
        }
      },
      { rootMargin: "200px" },
    );

    observer.observe(container);
    return () => observer.disconnect();
  }, [isActive, shouldLoad]);

  return (
    <div ref={containerRef} className="mb-3" style={{ minHeight: "450px" }}>
      {shouldLoad ? (
        <iframe
          className="w-100 border"
          style={{ minHeight: "450px" }}
          src={src}
          allowFullScreen=""
          loading="lazy"
          referrerPolicy="no-referrer-when-downgrade"
          title={title}
        />
      ) : (
        <div
          className="w-100 border d-flex align-items-center justify-content-center text-muted"
          style={{ minHeight: "450px" }}
        >
          Loading map…
        </div>
      )}
    </div>
  );
};

export default LazyMapIframe;
