<script>
  import { onMount } from 'svelte';

  let { runTool, bridge } = $props();

  let text = $state('');
  let format = $state('mp3');

  // Voices are loaded from the server's list_voices so the picker always
  // matches the active TTS provider (Gemini vs OpenAI). Fall back to a small
  // Gemini set if the call fails (Gemini is the default provider).
  let voices = $state([
    { id: 'Erinome', label: 'Erinome', desc: 'Clear' },
    { id: 'Puck', label: 'Puck', desc: 'Upbeat' },
    { id: 'Kore', label: 'Kore', desc: 'Firm' },
    { id: 'Aoede', label: 'Aoede', desc: 'Breezy' },
  ]);
  let voice = $state('Erinome');
  let provider = $state('');

  onMount(async () => {
    try {
      const res = await bridge.callTool('list_voices', {});
      const out = res?.structuredContent;
      if (out?.voices?.length) {
        voices = out.voices.map((v) => ({ id: v.id, label: v.label, desc: v.description }));
        voice = out.defaultVoice || voices[0].id;
        provider = out.provider || '';
      }
    } catch {
      /* keep the fallback list */
    }
  });

  async function handleSubmit(e) {
    e.preventDefault();
    if (!text.trim()) return;
    await runTool('synthesize_speech', {
      text,
      voice,
      responseFormat: format,
    });
  }
</script>

<form class="synthesize-form" onsubmit={handleSubmit}>
  <div class="field">
    <label for="syn-text">Text to Synthesize</label>
    <textarea
      id="syn-text"
      bind:value={text}
      placeholder="Type or paste text here. Use [PAUSE:5] for a 5-second pause..."
      rows="6"
      required
    ></textarea>
    <span class="hint">
      Tip: Use [PAUSE:N] markers to insert N-second silences
    </span>
  </div>

  <div class="row">
    <div class="field">
      <label for="syn-voice">Voice{#if provider}<span class="provider">· {provider}</span>{/if}</label>
      <select id="syn-voice" class="voice-select" bind:value={voice}>
        {#each voices as v}
          <option value={v.id}>{v.label}{v.desc ? ` — ${v.desc}` : ''}</option>
        {/each}
      </select>
    </div>

    <div class="field">
      <label for="syn-format">Format</label>
      <div class="format-toggle">
        <button
          type="button"
          class="format-btn"
          class:active={format === 'mp3'}
          onclick={() => format = 'mp3'}
        >
          MP3
        </button>
        <button
          type="button"
          class="format-btn"
          class:active={format === 'wav'}
          onclick={() => format = 'wav'}
        >
          WAV
        </button>
      </div>
    </div>
  </div>

  <div class="actions">
    <div class="char-count">
      {text.length > 0 ? `${text.replace(/\[PAUSE:\d+\]/g, '').length} chars` : ''}
    </div>
    <button type="submit" class="btn-primary" disabled={!text.trim()}>
      Synthesize Speech
    </button>
  </div>
</form>

<style>
  .synthesize-form {
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  label {
    font-size: 12px;
    font-weight: 600;
    color: #4a4a4a;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  textarea {
    padding: 14px;
    border: 1px solid #e8e5e0;
    border-radius: 12px;
    font-size: 15px;
    font-family: inherit;
    background: #fff;
    color: #2d2d2d;
    resize: vertical;
    min-height: 140px;
    line-height: 1.6;
    transition: border-color 0.2s, box-shadow 0.2s;
  }

  textarea:focus {
    outline: none;
    border-color: #8b7355;
    box-shadow: 0 0 0 3px rgba(139, 115, 85, 0.1);
  }

  textarea::placeholder {
    color: #b0a89e;
  }

  .hint {
    font-size: 11px;
    color: #999;
    font-style: italic;
  }

  .row {
    display: grid;
    grid-template-columns: 1fr;
    gap: 16px;
  }

  @media (min-width: 480px) {
    .row {
      grid-template-columns: 2fr 1fr;
    }
  }

  .provider {
    margin-left: 6px;
    font-weight: 500;
    color: #8b7355;
    text-transform: none;
    letter-spacing: 0;
  }

  .voice-select {
    padding: 11px 12px;
    border: 1px solid #e8e5e0;
    border-radius: 10px;
    background: #fff;
    font-size: 14px;
    font-family: inherit;
    color: #2d2d2d;
    cursor: pointer;
  }

  .voice-select:focus {
    outline: none;
    border-color: #8b7355;
    box-shadow: 0 0 0 3px rgba(139, 115, 85, 0.1);
  }

  .format-toggle {
    display: flex;
    gap: 4px;
    background: #f5f3f0;
    padding: 3px;
    border-radius: 8px;
  }

  .format-btn {
    flex: 1;
    padding: 10px;
    border: none;
    background: transparent;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 600;
    color: #6b6b6b;
    cursor: pointer;
    transition: all 0.2s;
  }

  .format-btn.active {
    background: #fff;
    color: #1a1a1a;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
  }

  .actions {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding-top: 4px;
  }

  .char-count {
    font-size: 12px;
    color: #999;
  }

  .btn-primary {
    padding: 12px 24px;
    background: linear-gradient(135deg, #8b7355 0%, #a08060 100%);
    color: #fff;
    border: none;
    border-radius: 10px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 2px 8px rgba(139, 115, 85, 0.25);
  }

  .btn-primary:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(139, 115, 85, 0.35);
  }

  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
