# Gateway API AI Policy

Gateway API is a project that values correctness, authenticity and community, and to maintain those values, we provide these guidelines on the use of Generative Artificial Intelligence ("Generative AI"), including Large Language Models (LLMs). In this document, these tools are referred to as "AI Tools".

In short, AI tools can be useful, and are acceptable in the project, as long as their use abides by this policy.

As a Kubernetes subproject, Gateway API is bound by the recommendations in the larger Kubernetes project, available in full [in the Kubernetes docs][kube-ai-guidance].

The most relevant parts for this document are listed here, further clarifications are listed later in this document:

* You must disclose the use of AI tools in the preparation of PRs or the initial opening of an issue.
* The use of AI tools in responding to comments on issues, PRs, discussions, or in any other means of communication with the project is not allowed.
* Large AI generated PRs and commit messages are not allowed.
* Listing AI tooling as a co-author, co-signing commits using an AI tool, or using the assisted-by, co-developed or similar commit trailer is not allowed.

The consequences for violations of this policy are described in the [AI Policy Violations](#ai-policy-violations) section.

As part of the CNCF, Kubernetes as a whole is also bound by the [Linux Foundation's policy on AI and copyright][lf-ai-policy], which can be broadly summarized for the purpose of this document as:

* Copyright for all submissions is assigned by the submitter to the project. AI tools cannot hold copyright.

## General principle

In addition to the rules above, we want to emphasize a guiding principle for all tool use, including AI tools:

**You are accountable for what tools do in your name**.

You, the human author behind a PR or issue, are responsible and accountable for the correctness, security and clarity of your contributions. Human review is required for all code and specifications before merging, so the community requests that you consider this and keep the number of code or specification changes that you open at once low. Starting out with smaller changes is good, as is keeping your requeted changes to one at a time.

## Acceptable Use

Generative AI can be a very useful tool, here are some broad categories of acceptable uses, each with a non-exhaustive list of examples:

* Anything that is only viewed by you. For example:
     * Use of a tool to summarize or understand the codebase or specification
     * Brainstorm ideas for code or specification updates (that is, GEPs), before you write the code or specification update yourself
     * Analyze your own submissions
     * You may use an AI tool to perform light copy-editing or grammar refinement of text you've written.
* Summarize existing facts as part of a design document. In this case, you are responsible for checking that all facts are correct, and that references exist. For example:
     * Preparing the Prior Art section of a GEP. You can use an AI tool to help gather the information about prior art, but you have to validate that all of that information is correct.
* Use AI tools to assist with preparing specification updates or code. For example:
    * You may use an AI tool to draft specification updates or code before you review, with the caveat that you must review every line, then revise and simplify as much as possible before submission.

These uses and other similar ones are acceptable as long as you:

* Are involved in the entire process of creating the contribution;
* Personally review, edit, and guarantee generated content before you submit it;
* Fully understand the content prior to submission;
* Take personal responsibility for the correctness, security and clarity of your submission;
* Report how you have used AI in your submission (we prefer to use the AI Influence Level or [AIL framework][ail-framework].

In particular, Gateway Enhancement Proposals (GEPs) and specification updates require _extreme_ precision in their language, something which is not the strong point of many AI tools. If you must use AI tools for these use cases, be aware the community expects an even higher level of review, editing and clarity for these documents than for other submissions.

GEPs in particular should generally not have a higher AIL than 2 (Human created, major AI augmentation).

### Labelling

Issues or PRs that are created with AI tool assistance will be marked with an `ai-assisted` label.

The purpose of this is _not_ to call out users of AI tools, it is to allow maintainers to gather data on how this policy is working, and if it needs any changes.

## Unacceptable Use

It is not acceptable to use AI tools to:

* Communicate in any Gateway API community space with content that is generated using AI tools. (Again, anything AIL:3 or above). That includes both any interactions on Github, Slack or in document editing apps.
  In particular:
    * If you analyze an issue or PR description using an AI tool, you MUST also write the description yourself and ensure that it is as concise as possible.
* Use an AI to write or rewrite comments for tone, style, or similar reasons. Authenticity is more important than awkward phrasing.
* Submit anything that you have not thoroughly reviewed and understood yourself, or submissions that you cannot explain or justify as part of a review.
* Rely solely on generative AI for problem solving, architecture, or document generation.

## Translation using AI tools

The Gateway API community understands that, for members who are not fluent in English, it can be intimidating to discuss complex technical topics in that language.

However, the community prefers the authenticity of communication that is not substantially written using AI tools.

* It is never acceptable to use AI tools without disclosing that use. This includes machine translation.
* Please use English to explain yourself first.
* If you feel that you are not being understood, you may use a translation tool, as long as you declare that use.

## AI Policy violations

If community participation is felt to not be abiding by this policy, members should expect to be asked by maintainers or other leadership if they are using AI tools, and that they should respond honestly.

However, there is one very important caveat to this:

* It is not acceptable to accuse a community member publicly of using AI tools based on style or other factors. If you do believe this may be happening, please reach out privately to the maintainers, so that they can have a discussion with the other community member.

This is not a situation where making a single mistake will have dire consequences, but contributors who misrepresent their use of AI tools are not welcome in this community.

Violations of this policy may result in consequences such as:

* PRs or issues being closed
* Contributors being banned from participating in the project.

[kube-ai-guidance]: https://www.kubernetes.dev/docs/guide/pull-requests/#ai-guidance
[ail-framework]: https://danielmiessler.com/blog/ai-influence-level-ail
[lf-ai-policy]: https://www.linuxfoundation.org/legal/generative-ai