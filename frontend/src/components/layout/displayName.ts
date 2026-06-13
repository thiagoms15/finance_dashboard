export function displayName(name: string | undefined, email: string | undefined) {
  if (name?.trim()) {
    return name.trim();
  }

  const localPart = email?.split("@")[0]?.trim();
  if (!localPart) {
    return "Investor";
  }

  return localPart
    .replace(/[._-]+/g, " ")
    .split(" ")
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}
