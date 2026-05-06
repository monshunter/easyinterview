import type { FC } from "react";

import type { Route } from "../routes";

/**
 * D1 placeholder for the User Profile route. Mirrors
 * docs/ui-design/user-profile-and-settings.md §3 sections without rendering
 * full evidence cards — D2-D6 owners populate the AI-summarized content.
 *
 * Critically: this shell never restores the retired Growth / Experiences /
 * Mistakes / Drill modules. The settings page handles字体预设; the profile
 * page does NOT carry job preferences, target role, or skill tags.
 */
export const ProfileScreen: FC<{ route: Route }> = ({ route }) => (
  <section
    data-testid={`route-${route.name}`}
    data-route-name={route.name}
    data-route-params={JSON.stringify(route.params)}
  >
    <header>
      <h1>用户画像</h1>
      <p>系统理解你的那份画像</p>
    </header>
    <div data-testid="profile-identity-summary">
      <h2>身份摘要</h2>
      <ul>
        <li>姓名 / 昵称</li>
        <li>职业摘要</li>
        <li>置信度</li>
        <li>来源统计：简历 / JD / 模拟 / 复盘</li>
      </ul>
    </div>
    <div data-testid="profile-sections">
      <h2>分区</h2>
      <ul>
        <li>职业定位</li>
        <li>技能与深度</li>
        <li>经历证据</li>
        <li>面试表现</li>
      </ul>
    </div>
    <div data-testid="profile-insight-cards">
      <h2>洞察卡片</h2>
      <p>AI 当前推断 / 置信度 / 来源 / 修正入口（D2-D6 接入）</p>
    </div>
    <div data-testid="profile-used-by">
      <h2>被哪些模块使用</h2>
      <ul>
        <li>岗位推荐</li>
        <li>模拟面试规划</li>
        <li>报告分析维度</li>
      </ul>
    </div>
    <div data-testid="profile-recent-evidence">
      <h2>最近证据来源</h2>
      <p>简历 / JD / 模拟 / 复盘 的近期更新（D2-D6 接入）</p>
    </div>
  </section>
);
