import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';

const FeatureList = [
  {
    title: '🔥 Ekspresivitas Tinggi',
    icon: '⚡',
    description: (
      <>
        Buat aturan kompleks dengan nested logic dan kondisi berlapis. 
        Mudah memodelkan edge-case logic seperti loop, short-circuiting, 
        atau nested evaluation untuk kebutuhan bisnis yang rumit.
      </>
    ),
  },
  {
    title: '🧩 Reusable & Modular',
    icon: '🔧',
    description: (
      <>
        Fungsi (fun) memungkinkan rule dipecah menjadi blok yang dapat digunakan ulang.
        Developer bisa menyusun ruleset libraries untuk berbagai domain aplikasi.
      </>
    ),
  },
  {
    title: '🚀 Flow Control Penuh',
    icon: '⚙️',
    description: (
      <>
        Rule tidak hanya "if this then that", tapi juga mencakup perulangan 
        (cek fraud di 10 transaksi terakhir), break/continue untuk early exit,
        dan nested rule evaluation yang kompleks.
      </>
    ),
  },
  {
    title: '🔍 Debuggable & Observable',
    icon: '�',
    description: (
      <>
        Karena bisa dieksekusi seperti program, Anda bisa pasang logging/debug print 
        di dalam rule untuk observabilitas dan troubleshooting yang mudah.
      </>
    ),
  },
  {
    title: '🎯 Deklaratif + Imperatif',
    icon: '📝',
    description: (
      <>
        Fleksibilitas menulis rule secara deklaratif (if this then that) atau 
        imperatif (run this logic if...), cocok untuk domain keamanan, 
        antifraud, dan automation.
      </>
    ),
  },
  {
    title: '🔄 Runtime Programmable',
    icon: '💾',
    description: (
      <>
        Simpan rule di database, load dan jalankan secara dinamis. 
        Modifikasi rule tanpa perlu recompile sistem utama untuk 
        adaptabilitas maksimal.
      </>
    ),
  },
];

function Feature({ icon, title, description }) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <div className={styles.featureIcon}>{icon}</div>
      </div>
      <div className="text--center padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
