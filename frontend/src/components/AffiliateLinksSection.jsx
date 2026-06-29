import useAffiliateLinks from "../hooks/useAffiliateLinks";
import AffiliateLinks from "./AffiliateLinks";

export default function AffiliateLinksSection() {
  const { links, isLoading, error } = useAffiliateLinks(true);

  if (isLoading) {
    return null;
  }

  return <AffiliateLinks links={links} error={error} />;
}
