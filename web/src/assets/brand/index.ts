import forumLogo from "./logo.png";
import forumBackground from "./background.png";

import userAvatar1 from "./avatars/user-1.png";
import userAvatar2 from "./avatars/user-2.png";
import userAvatar3 from "./avatars/user-3.png";
import userAvatar4 from "./avatars/user-4.png";
import userAvatar5 from "./avatars/user-5.png";
import userAvatar6 from "./avatars/user-6.png";

import aiIcon from "./icons/ai.png";
import analyticsIcon from "./icons/analytics.png";
import composeIcon from "./icons/compose.png";
import documentSearchIcon from "./icons/document-search.png";
import hotTagIcon from "./icons/hot-tag.png";
import notificationIcon from "./icons/notification.png";
import profileIcon from "./icons/profile.png";
import replyIcon from "./icons/reply.png";
import searchIcon from "./icons/search.png";
import securityIcon from "./icons/security.png";
import settingsIcon from "./icons/settings.png";
import trendIcon from "./icons/trend.png";

export { forumBackground, forumLogo };

export const defaultUserAvatars = [
  userAvatar1,
  userAvatar2,
  userAvatar3,
  userAvatar4,
  userAvatar5,
  userAvatar6,
];

export function defaultUserAvatar(seed: string | number): string {
  const text = String(seed);
  let hash = 0;
  for (let i = 0; i < text.length; i += 1) {
    hash = (hash * 31 + text.charCodeAt(i)) >>> 0;
  }
  return defaultUserAvatars[hash % defaultUserAvatars.length];
}

export const brandIcons = {
  ai: aiIcon,
  analytics: analyticsIcon,
  compose: composeIcon,
  documentSearch: documentSearchIcon,
  hotTag: hotTagIcon,
  notification: notificationIcon,
  profile: profileIcon,
  reply: replyIcon,
  search: searchIcon,
  security: securityIcon,
  settings: settingsIcon,
  trend: trendIcon,
};

export type BrandIconName = keyof typeof brandIcons;
