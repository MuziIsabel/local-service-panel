---
description: 执行 TODO.md 的下一项任务，完成后按项目标准自检并修复问题
---
请在当前项目中工作。

请先阅读：
- AGENTS.md
- TODO.md
- CONTEXT.md
- README.md
- docs/WORKFLOW.md
- docs/DEFINITION_OF_DONE.md
- docs/TESTING.md
- docs/DEV_ENV.md
- docs/SECURITY.md
- 与本轮任务相关的 docs/specs/

然后执行 TODO.md 中的下一项任务。

要求：
1. 只做当前下一项任务，不要扩展到后续功能；
2. 遵循 AGENTS.md、docs/WORKFLOW.md、docs/DEFINITION_OF_DONE.md；
3. 如果涉及具体功能规格，优先阅读 docs/specs/；
4. 完成实现后运行可行验证；
5. 同步更新 TODO.md；
6. 如果影响 API/数据模型/安全/架构/开发命令，同步更新相关 docs；
7. 完成后按下面的“自检规则”进行自我 review；
8. 如果自检发现问题，先修复问题，再重新检查；
9. 最终汇报时必须包含自检结果。

自检规则：
1. 当前改动是否符合本轮任务范围；
2. 是否有过度设计或无关改动；
3. 是否违反安全约束；
4. API/数据模型/架构/安全/开发命令变化是否同步了 docs；
5. 是否运行了可行验证，或说明了不能验证的原因；
6. TODO.md 是否正确更新；
7. 是否满足 docs/DEFINITION_OF_DONE.md。

最后按以下格式汇报：

完成内容：
- ...

修改文件：
- ...

验证结果：
- ...

TODO 更新：
- ...

自检结果：
- ...

风险/注意：
- ...

下一步建议：
- ...
