---
title: MalZone Documentation
description: Architecture, implementation truth, and operating design for enterprise-controlled malware analysis.
home: true
---

<section class="hero">
  <div class="hero-copy">
    <div class="eyebrow">Enterprise-controlled analysis</div>
    <h1>Keep malware analysis <span>inside your boundary.</span></h1>
    <p class="hero-lead">MalZone is a self-hosted, API-first platform for safely submitting, observing, interacting with, reporting on, and destroying disposable Windows analysis environments—without surrendering samples, evidence, identity, or operating control.</p>
    <div class="hero-actions">
      <a class="button button-primary" href="design/high-level/README.md">Explore the architecture</a>
      <a class="button" href="design/high-level/00-implementation-conformance.md">See implementation truth</a>
    </div>
  </div>
  <div class="boundary-card" aria-label="MalZone control boundary">
    <p class="boundary-label">Control boundary / local</p>
    <div class="boundary-row"><span class="boundary-index">01</span><div><strong>Submit</strong><span>Quarantine, hash, admit, and bind every input to policy.</span></div></div>
    <div class="boundary-row"><span class="boundary-index">02</span><div><strong>Observe</strong><span>Live behavior, evidence, network activity, and collection health.</span></div></div>
    <div class="boundary-row"><span class="boundary-index">03</span><div><strong>Interact</strong><span>Human or bounded agent actions through one governed API.</span></div></div>
    <div class="boundary-row"><span class="boundary-index">04</span><div><strong>Destroy</strong><span>No terminal result until disposable resources and credentials are gone.</span></div></div>
    <div class="boundary-rule">Public cloud dependency: none required</div>
  </div>
</section>

<section class="signal-strip" aria-label="Architecture principles">
  <div class="signal-grid">
    <div class="signal-item"><div class="signal-code">LOCAL_01</div><strong>Local by design</strong><span>Air-gapped operation remains a baseline, not an enterprise add-on.</span></div>
    <div class="signal-item"><div class="signal-code">API_02</div><strong>API-first</strong><span>The UI, analysts, agents, and integrations share governed contracts.</span></div>
    <div class="signal-item"><div class="signal-code">VM_03</div><strong>Disposable execution</strong><span>Fresh identity, storage, and networks for every analysis.</span></div>
    <div class="signal-item"><div class="signal-code">TRUTH_04</div><strong>Evidence over claims</strong><span>Conformance distinguishes what exists from what is only designed.</span></div>
  </div>
</section>

<section class="home-section">
  <div class="section-heading">
    <span class="section-number">01 / EXPLORE</span>
    <div class="section-heading-copy"><h2>The system, from strategy to containment.</h2><p>Start with the outcome you care about. Every design area links back to implementation truth, security boundaries, and executable acceptance evidence.</p></div>
  </div>
  <div class="feature-grid">
    <div class="feature-card"><span class="feature-code">ARCH / 001</span><h3>Architecture and lifecycle</h3><p>Control, analysis, data, image-supply, integration, and operations planes—joined by a deterministic disposable-session lifecycle.</p><a href="design/high-level/design_01.md">Open architecture</a></div>
    <div class="feature-card"><span class="feature-code">AUTO / 002</span><h3>AI automation and SIEM</h3><p>Agents propose closed-schema actions. Deterministic policy executes them. Credential-isolated adapters export disclosure-controlled events.</p><a href="design/high-level/10-overall/07-ai-automation-siem.md">Review automation design</a></div>
    <div class="feature-card"><span class="feature-code">IMAGE / 003</span><h3>Custom Windows environments</h3><p>Exact software manifests resolve into immutable, reviewed Windows images without runtime Internet installation.</p><a href="design/high-level/10-overall/05-software-catalog-image-composition.md">See image composition</a></div>
    <div class="feature-card"><span class="feature-code">ZERO / 004</span><h3>Zero-trust containment</h3><p>The Windows guest is assumed compromised. Network, identity, relay, artifact, and cleanup controls are designed from that premise.</p><a href="design/high-level/30-security/01-threat-model-zero-trust.md">Read the threat model</a></div>
  </div>
</section>

<section class="home-section">
  <div class="section-heading">
    <span class="section-number">02 / REALITY</span>
    <div class="section-heading-copy"><h2>Working software, without inflated claims.</h2><p>The harmless Kubernetes POC proves lifecycle and integration plumbing. It deliberately does not present itself as a Windows malware sandbox.</p></div>
  </div>
  <div class="reality-grid">
    <div class="reality-panel implemented">
      <span class="panel-state">Implemented POC</span>
      <h3>Control-plane spine</h3>
      <p>Executable today in the local k3d development environment.</p>
      <ul><li>Go API, Analysis CRD, and reconciler</li><li>Restricted tokenless canary runner</li><li>Observe-only agent action with shell denial</li><li>Metadata-only ECS adapter and test sink</li><li>Helm packaging and live cleanup evidence</li></ul>
    </div>
    <div class="reality-panel target">
      <span class="panel-state">Target architecture</span>
      <h3>Production analysis plane</h3>
      <p>Designed with acceptance gates; not claimed as shipped.</p>
      <ul><li>Disposable Windows/KubeVirt clone</li><li>Dual isolated networks and session relay</li><li>Live console and behavior collectors</li><li>Quarantined evidence and deterministic reports</li><li>OIDC, durable data services, and supported SIEM</li></ul>
    </div>
  </div>
</section>

<section class="home-section architecture-visual">
  <div class="section-heading">
    <span class="section-number">03 / FLOW</span>
    <div class="section-heading-copy"><h2>One evidence-backed design chain.</h2><p>Business requirements become architecture decisions, machine-readable contracts, implementation evidence, and an honest conformance status.</p></div>
  </div>
  <pre><code class="language-mermaid">flowchart LR
    Business["Business need"] --&gt; Design["Architecture"]
    Design --&gt; ADR["Decisions"]
    Design --&gt; Contracts["Contracts"]
    ADR &amp; Contracts --&gt; Build["Implementation"]
    Build --&gt; Evidence["Tests + evidence"]
    Evidence --&gt; Truth["Conformance"]
    Truth -. "gap" .-&gt; Design</code></pre>
</section>

<section class="home-section">
  <div class="section-heading">
    <span class="section-number">04 / GOVERN</span>
    <div class="section-heading-copy"><h2>Architecture stays synchronized with code.</h2><p>Human and AI-assisted changes follow the same major-change policy, contract discipline, evidence requirements, and changed-file design gate.</p></div>
  </div>
  <div class="governance-grid">
    <a data-code="RULES" href="../CLAUDE.md">Canonical repository rules</a>
    <a data-code="AGENTS" href="../AGENTS.md">Coding-agent discovery</a>
    <a data-code="CONTRIB" href="../CONTRIBUTING.md">Human contribution guide</a>
    <a data-code="SYNC" href="prompts/governance/README.md">Major-change governance</a>
  </div>
  <div class="publication-note">This portal publishes only an explicit allow-list of committed design, contract, governance, and development files. It excludes deployment-private supplements, credentials, samples, evidence, internal endpoints, and malware artifacts.</div>
</section>
