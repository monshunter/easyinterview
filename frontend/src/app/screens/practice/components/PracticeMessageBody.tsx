import type { FC } from "react";
import Markdown, { defaultUrlTransform, type UrlTransform } from "react-markdown";
import remarkGfm from "remark-gfm";

export interface PracticeMessageBodyProps {
  text: string;
}

const allowedSchemes = new Set(["http", "https", "mailto"]);
const schemePattern = /^([a-z][a-z\d+.-]*):/i;

const safePracticeUrlTransform: UrlTransform = (url) => {
  const safeUrl = defaultUrlTransform(url);
  if (!safeUrl || safeUrl.startsWith("//")) return "";

  const scheme = safeUrl.match(schemePattern)?.[1]?.toLowerCase();
  return !scheme || allowedSchemes.has(scheme) ? safeUrl : "";
};

const isExternalHttpUrl = (url: string): boolean => /^https?:\/\//i.test(url);

export const PracticeMessageBody: FC<PracticeMessageBodyProps> = ({ text }) => (
  <div data-testid="practice-message-body" className="ei-practice-message-body">
    <Markdown
      remarkPlugins={[remarkGfm]}
      skipHtml
      urlTransform={safePracticeUrlTransform}
      components={{
        img: () => null,
        a: ({ href, children }) => {
          if (!href) return <span>{children}</span>;
          const external = isExternalHttpUrl(href);
          return (
            <a
              href={href}
              target={external ? "_blank" : undefined}
              rel={external ? "noopener noreferrer" : undefined}
            >
              {children}
            </a>
          );
        },
      }}
    >
      {text}
    </Markdown>
  </div>
);
