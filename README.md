Fetcher/Extractor data pergitungan suara (sirekap) KPU untuk pemilu 2024. Dirancang untuk dijalankan secara berkala melaului github action (lihat `.github/workflows/cli-fetch-votes.yml`).

> :warning: **Gunakan dengan Bijaksana:** Sebelum melakukan scrapper ke web manapun, pastikan Anda memahami dan mematuhi praktik web scraping yang etis. Hormati situs web yang Anda scrap dan patuhi syarat penggunaannya.

### Disclaimer:
Scraper ini disediakan apa adanya, tanpa jaminan atau garansi. Penulis tidak bertanggung jawab atas penggunaan yang salah atau tidak sah dari kode ini.

---

### Catatan:
Proyek ini dibuat dalam rangka mempelajari cara menggunakan dan memanfaatkan fitur GitHub Actions untuk otomatisasi tugas-tugas tertentu dalam pengembangan perangkat lunak.

### Cara Kerja
1. Github action run secara berkala tiap jam: 
   ```yml
   on:
    ...
    schedule:
      - cron: '0 * * * *'
   ...
   ```
2. Eksekusi CLI
   ```yml
   ...
    jobs:
      build:
        runs-on: ubuntu-latest
    
        steps:
          ...
          - name: Execute CLI fetchVotes
            run: go run presenter/cli/main.go fetchVotes
          ...
   ```
  
3. fetch data perhitungan presiden dan partai dari website KPU (lihat `presenter/cli/cmd/fetchVotes.go`)
4. extract data dan simpan ke file csv ()
   ```yml
   jobs:
      build:
        runs-on: ubuntu-latest
    
        steps:
          ...
          - name: Commit output file
            run: |
              git config --local user.email "action@github.com"
              git config --local user.name "GitHub Action"
              git add .
              git commit -m "fetc votes - $(date -u +"%Y-%m-%d %H:%M:%S UTC")"
              git push
   ```

   
