# EnvKit 桌面客户端视觉规范

## 设计哲学

> "像工具一样思考，而非像产品"

### 核心原则
1. **信息密度优先** - 一屏展示尽可能多的有用信息
2. **减法设计** - 去掉所有装饰性元素
3. **系统原生感** - 像系统自带应用一样自然
4. **功能可见性** - 所有操作清晰可见，无需探索

---

## 配色系统

### Light Mode（默认主题）

```css
/* 基础色板 */
:root {
  /* 背景层级 */
  --bg-primary:    #ffffff;   /* 主背景 */
  --bg-secondary:  #f5f5f7;   /* 侧边栏背景 */
  --bg-tertiary:   #efefef;   /* 卡片背景 */
  --bg-hover:      #e8e8e8;   /* 悬停状态 */
  --bg-active:     #d1d1d6;   /* 激活状态 */
  
  /* 文本层级 */
  --text-primary:   #1d1d1f;  /* 主标题、正文 */
  --text-secondary: #6e6e73;  /* 次要信息 */
  --text-tertiary:  #86868b;  /* 辅助说明 */
  --text-disabled:  #c7c7cc;  /* 禁用状态 */
  
  /* 边框 */
  --border-light:   #e5e5e5;  /* 轻边框 */
  --border-medium:  #d1d1d6;  /* 标准边框 */
  --border-heavy:   #a1a1a6;  /* 重边框 */
  
  /* 功能色 */
  --color-success:  #28a745;  /* 成功、已安装 */
  --color-warning:  #ff9500;  /* 警告、需更新 */
  --color-error:    #dc3545;  /* 错误、失败 */
  --color-info:     #007aff;  /* 信息、链接 */
  --color-neutral:  #8e8e93;  /* 中性 */
  
  /* 功能色背景（10% 透明度）*/
  --color-success-bg: #28a74510;
  --color-warning-bg: #ff950010;
  --color-error-bg:   #dc354510;
  --color-info-bg:    #007aff10;
}
```

### Dark Mode（可选）

```css
:root[data-theme="dark"] {
  /* 背景层级 */
  --bg-primary:    #1e1e1e;
  --bg-secondary:  #2d2d2d;
  --bg-tertiary:   #3a3a3a;
  --bg-hover:      #4a4a4a;
  --bg-active:     #5a5a5a;
  
  /* 文本层级 */
  --text-primary:   #ffffff;
  --text-secondary: #adadb8;
  --text-tertiary:  #86868b;
  --text-disabled:  #636366;
  
  /* 边框 */
  --border-light:   #3e3e42;
  --border-medium:  #48484a;
  --border-heavy:   #636366;
  
  /* 功能色保持不变 */
}
```

---

## 字体系统

### 字体栈
```css
/* 系统原生字体 */
font-family: 
  -apple-system,              /* macOS/iOS */
  BlinkMacSystemFont,         /* macOS Chrome */
  "Segoe UI",                 /* Windows */
  "Roboto",                   /* Android */
  "Helvetica Neue",           /* macOS fallback */
  Arial,                      /* Universal fallback */
  sans-serif;
```

### 字体尺寸
```css
--font-xxl:  24px;   /* 页面标题 */
--font-xl:   18px;   /* 区块标题 */
--font-lg:   15px;   /* 卡片标题 */
--font-base: 13px;   /* 正文、按钮 */
--font-sm:   11px;   /* 辅助信息 */
--font-xs:   10px;   /* 标签、角标 */
```

### 字重
```css
--font-regular: 400;
--font-medium:  500;
--font-semibold: 600;
```

---

## 间距系统

### 基础单位：4px
```css
--space-1:  4px;
--space-2:  8px;
--space-3:  12px;
--space-4:  16px;
--space-5:  20px;
--space-6:  24px;
--space-8:  32px;
--space-10: 40px;
```

### 应用规则
- **组件内间距**: 12-16px
- **组件间间距**: 16-24px
- **区块间间距**: 24-32px
- **页面边距**: 20px

---

## 圆角系统

```css
--radius-none: 0px;
--radius-sm:   4px;   /* 按钮、输入框 */
--radius-md:   6px;   /* 卡片 */
--radius-lg:   8px;   /* 弹窗 */
--radius-full: 999px; /* 标签、徽章 */
```

### 使用建议
- **小组件**（按钮、输入框）: 4px
- **卡片**: 6px
- **弹窗**: 8px
- **避免过度圆角**（>12px）

---

## 阴影系统

### 极简阴影（仅在必要时使用）
```css
/* 悬浮卡片 */
--shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.08);

/* 弹窗 */
--shadow-md: 0 4px 12px rgba(0, 0, 0, 0.12);

/* 模态框 */
--shadow-lg: 0 8px 24px rgba(0, 0, 0, 0.15);
```

### 使用原则
- **默认不用阴影**，用边框代替
- **只在需要层级关系时使用**
- **Dark Mode 下阴影更重**

---

## 组件规范

### 1. 按钮

#### 主按钮（Primary）
```css
.btn-primary {
  background: var(--color-info);
  color: white;
  padding: 6px 16px;
  border-radius: 4px;
  font-size: 13px;
  font-weight: 500;
  border: none;
  cursor: pointer;
}

.btn-primary:hover {
  opacity: 0.9;
}

.btn-primary:active {
  opacity: 0.8;
}
```

#### 次要按钮（Secondary）
```css
.btn-secondary {
  background: var(--bg-tertiary);
  color: var(--text-primary);
  padding: 6px 16px;
  border-radius: 4px;
  font-size: 13px;
  border: 1px solid var(--border-light);
  cursor: pointer;
}
```

#### 危险按钮（Danger）
```css
.btn-danger {
  background: var(--color-error);
  color: white;
  /* 其他样式同 primary */
}
```

#### 尺寸变体
```css
/* Small */
.btn-sm {
  padding: 4px 12px;
  font-size: 12px;
}

/* Large */
.btn-lg {
  padding: 8px 20px;
  font-size: 14px;
}
```

### 2. 输入框

```css
.input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--border-medium);
  border-radius: 4px;
  font-size: 13px;
  background: var(--bg-primary);
  color: var(--text-primary);
}

.input:focus {
  outline: none;
  border-color: var(--color-info);
  box-shadow: 0 0 0 3px var(--color-info-bg);
}

.input:disabled {
  background: var(--bg-secondary);
  color: var(--text-disabled);
  cursor: not-allowed;
}
```

### 3. 卡片

```css
.card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-light);
  border-radius: 6px;
  padding: 16px;
}

.card:hover {
  border-color: var(--border-medium);
}
```

### 4. 标签（Badge）

```css
.badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 500;
}

.badge-success {
  background: var(--color-success-bg);
  color: var(--color-success);
}

.badge-warning {
  background: var(--color-warning-bg);
  color: var(--color-warning);
}

.badge-error {
  background: var(--color-error-bg);
  color: var(--color-error);
}
```

### 5. 表格

```css
.table {
  width: 100%;
  border-collapse: collapse;
}

.table th {
  text-align: left;
  padding: 8px 12px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid var(--border-medium);
}

.table td {
  padding: 12px;
  font-size: 13px;
  border-bottom: 1px solid var(--border-light);
}

.table tr:hover {
  background: var(--bg-hover);
}
```

---

## 图标系统

### 图标库选择
**推荐：Lucide Icons**
- SVG 格式
- 24x24 设计尺寸
- 2px 描边宽度
- 一致的视觉风格

### 使用规范
```css
.icon {
  width: 16px;
  height: 16px;
  stroke: currentColor;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.icon-sm { width: 14px; height: 14px; }
.icon-lg { width: 20px; height: 20px; }
```

### 常用图标
```
✓ 成功：check-circle
✗ 失败：x-circle
⚠ 警告：alert-triangle
ℹ 信息：info
⚙ 设置：settings
+ 添加：plus
− 删除：trash-2
↻ 刷新：refresh-cw
▶ 运行：play
■ 停止：square
```

---

## 动画规范

### 过渡时间
```css
--duration-fast:   100ms;  /* 悬停效果 */
--duration-normal: 200ms;  /* 一般过渡 */
--duration-slow:   300ms;  /* 复杂动画 */
```

### 缓动函数
```css
--ease-out: cubic-bezier(0.25, 0.1, 0.25, 1);
--ease-in-out: cubic-bezier(0.42, 0, 0.58, 1);
```

### 使用原则
- **最小化动画使用**
- **只动画 opacity 和 transform**（性能最优）
- **避免动画 height, width, top, left**

---

## 布局规范

### 侧边栏
```
宽度：200px（固定）
背景：var(--bg-secondary)
边框：1px solid var(--border-light)（右侧）
内边距：16px 0
```

### 主内容区
```
最大宽度：1200px
内边距：20px
```

### 网格系统
```css
/* 2列网格 */
.grid-2 {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

/* 3列网格 */
.grid-3 {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

/* 响应式 */
@media (max-width: 768px) {
  .grid-3 {
    grid-template-columns: repeat(2, 1fr);
  }
}
```

---

## 响应式断点

```css
/* 平板 */
@media (max-width: 1024px) {
  /* 侧边栏可折叠 */
}

/* 手机 */
@media (max-width: 768px) {
  /* 侧边栏变为顶部导航 */
}
```

---

## 可访问性

### 颜色对比度
- **正文文本**: 至少 4.5:1
- **大号文本**: 至少 3:1
- **UI 组件**: 至少 3:1

### 焦点指示
```css
:focus-visible {
  outline: 2px solid var(--color-info);
  outline-offset: 2px;
}
```

### 键盘导航
- 所有交互元素可通过 Tab 访问
- 支持 Enter/Space 激活
- 支持 Esc 关闭弹窗

---

## 平台差异化

### macOS
```css
/* 使用 SF Pro 字体 */
font-family: -apple-system, BlinkMacSystemFont;

/* 更小的圆角 */
border-radius: 4px;

/* 更细的边框 */
border-width: 0.5px;
```

### Windows
```css
/* 使用 Segoe UI 字体 */
font-family: "Segoe UI";

/* 更方正的风格 */
border-radius: 2px;

/* 标准边框 */
border-width: 1px;
```

### Linux
```css
/* 使用系统字体 */
font-family: system-ui;

/* 遵循 GTK 风格 */
```

---

## 设计检查清单

在完成 UI 设计后，检查以下项目：

- [ ] 所有文本至少 11px
- [ ] 颜色对比度符合 WCAG AA 标准
- [ ] 所有交互元素有悬停/激活状态
- [ ] 无多余装饰元素
- [ ] 信息密度合理（不过于稀疏）
- [ ] 间距一致（使用 4px 基础单位）
- [ ] 圆角不超过 8px
- [ ] 阴影使用克制
- [ ] 动画不超过 300ms
- [ ] 支持键盘导航

---

## 参考资源

### 设计灵感
- **Apple Human Interface Guidelines**
- **Microsoft Fluent Design System**
- **Linear Design**
- **Sublime Text**

### 工具推荐
- **Figma**: UI 设计
- **Lucide Icons**: 图标库
- **Coolors**: 配色工具
- **WebAIM**: 对比度检查
