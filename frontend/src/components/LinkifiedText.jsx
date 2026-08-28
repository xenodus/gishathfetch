import { splitTextWithLinks } from "../utils/linkifyText";

const LinkifiedText = ({ text }) => {
  const parts = splitTextWithLinks(text);

  return (
    <>
      {parts.map((part) =>
        part.type === "link" ? (
          <a
            key={`link-${part.start}`}
            href={part.href}
            target="_blank"
            rel="noopener noreferrer"
          >
            {part.value}
          </a>
        ) : (
          <span key={`text-${part.start}`}>{part.value}</span>
        ),
      )}
    </>
  );
};

export default LinkifiedText;
