import React from 'react';

// src内にある場合は、このようにimportする必要があります
import correctWav from '../sounds/決定ボタンを押す52.mp3';
import incorrectWav from '../sounds/ビープ音4.mp3';
import hotspotWav from '../sounds/決定ボタンを押す37.mp3';

const sounds = {
  correct: new Audio(correctWav),
  incorrect: new Audio(incorrectWav),
  hostspot: new Audio(hotspotWav),
};

export const playSE = (key) => {
  const audio = sounds[key];
  if (audio) {
    audio.currentTime = 0; // 連続再生対応
    audio.play().catch(e => console.log("再生ブロックまたはエラー:", e));
  }
};