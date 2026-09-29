const siteApiPaths = [
  /^loadpoints\//,
  /^vehicles\//,
  /^(buffer|battery|grid|priority|profile|residual|solar|smart|optimizer)/,
  /^tariff\//,
  /^sessions(?:[/?]|$)/,
  /^session\//,
  /^gridsessions(?:[/?]|$)/,
  /^history\//,
  /^optimize(?:[/?]|$)/,
];

let activeSite: string | null = null;

export function setActiveSite(site: string | null) {
  activeSite = site;
}

export function getActiveSite() {
  return activeSite;
}

export function siteApiUrl(url: string, site: string | null) {
  const path = url.replace(/^\//, "");
  if (site && /^db\/metrics(?:[/?]|$)/.test(path)) {
    const separator = path.includes("?") ? "&" : "?";
    return `${url}${separator}site=${encodeURIComponent(site)}`;
  }
  return site && siteApiPaths.some((pattern) => pattern.test(path))
    ? `sites/${encodeURIComponent(site)}/${path}`
    : url;
}
