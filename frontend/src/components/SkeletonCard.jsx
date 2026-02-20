import { Col } from "react-bootstrap";

export default function SkeletonCard() {
  return (
    <Col xs={6} lg={3} className="mb-4">
      <div className="text-center mb-2">
        <div
          className="skeleton-image bg-secondary-subtle rounded w-100"
          style={{
            aspectRatio: "304/424",
            animation: "pulse 1.5s infinite ease-in-out",
          }}
        />
      </div>
      <div className="text-center">
        <div
          className="skeleton-text bg-secondary-subtle rounded mb-1 mx-auto"
          style={{
            height: "1rem",
            width: "80%",
            animation: "pulse 1.5s infinite ease-in-out",
          }}
        />
        <div
          className="skeleton-text bg-secondary-subtle rounded mb-1 mx-auto"
          style={{
            height: "0.8rem",
            width: "40%",
            animation: "pulse 1.5s infinite ease-in-out",
          }}
        />
        <div
          className="skeleton-text bg-secondary-subtle rounded mb-2 mx-auto"
          style={{
            height: "0.8rem",
            width: "60%",
            animation: "pulse 1.5s infinite ease-in-out",
          }}
        />
        <div
          className="skeleton-button bg-primary-subtle rounded mx-auto"
          style={{
            height: "2rem",
            width: "50%",
            animation: "pulse 1.5s infinite ease-in-out",
          }}
        />
      </div>
    </Col>
  );
}
