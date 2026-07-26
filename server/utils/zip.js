/**
 * 纯 Node.js ZIP 工具（零依赖）
 *
 * 只支持"存储"模式（无压缩），对插件这种小文本文件够用了
 * 生成的 .zip 文件可以被 Windows 资源管理器、7-Zip、WinRAR 正常打开
 */

const fs = require('fs')
const path = require('path')

/**
 * 将目录打包为 zip 文件
 * @param {string} dirPath - 要打包的目录
 * @param {string} zipPath - 输出的 zip 文件路径
 */
function packDir(dirPath, zipPath) {
  const files = collectFiles(dirPath)
  const buf = createZipBuffer(files, dirPath)
  fs.writeFileSync(zipPath, buf)
}

/**
 * 收集目录下所有文件（递归）
 */
function collectFiles(dirPath) {
  const result = []
  const walk = (dir, relPath) => {
    const entries = fs.readdirSync(dir, { withFileTypes: true })
    for (const entry of entries) {
      const fullPath = path.join(dir, entry.name)
      const zipPath = relPath ? relPath + '/' + entry.name : entry.name
      if (entry.isDirectory()) {
        walk(fullPath, zipPath)
      } else {
        result.push({ zipPath, fullPath, data: fs.readFileSync(fullPath) })
      }
    }
  }
  walk(dirPath, '')
  return result
}

/**
 * 生成 ZIP 格式的 Buffer
 *
 * ZIP 文件结构：
 *   [Local File Header + File Data] × N
 *   [Central Directory Entry] × N
 *   [End of Central Directory Record]
 */
function createZipBuffer(files, baseDir) {
  const encoder = new TextEncoder()
  let localHeaders = Buffer.alloc(0)
  let centralDir = Buffer.alloc(0)
  let offset = 0

  for (const file of files) {
    const nameBytes = encoder.encode(file.zipPath)
    const data = file.data
    const crc = crc32(data)

    // --- Local File Header ---
    const localHeader = Buffer.alloc(30 + nameBytes.length)
    let pos = 0
    localHeader.writeUInt32LE(0x04034b50, pos); pos += 4  // signature
    localHeader.writeUInt16LE(20, pos); pos += 2            // version needed
    localHeader.writeUInt16LE(0, pos); pos += 2             // flags
    localHeader.writeUInt16LE(0, pos); pos += 2             // compression: 0=store
    localHeader.writeUInt16LE(0, pos); pos += 2             // mod time
    localHeader.writeUInt16LE(0, pos); pos += 2             // mod date
    localHeader.writeUInt32LE(crc, pos); pos += 4           // crc32
    localHeader.writeUInt32LE(data.length, pos); pos += 4   // compressed size
    localHeader.writeUInt32LE(data.length, pos); pos += 4   // uncompressed size
    localHeader.writeUInt16LE(nameBytes.length, pos); pos += 2  // filename length
    localHeader.writeUInt16LE(0, pos); pos += 2             // extra field length
    Buffer.from(nameBytes).copy(localHeader, pos)

    // --- Central Directory Entry ---
    const centEntry = Buffer.alloc(46 + nameBytes.length)
    pos = 0
    centEntry.writeUInt32LE(0x02014b50, pos); pos += 4      // signature
    centEntry.writeUInt16LE(20, pos); pos += 2               // version made by
    centEntry.writeUInt16LE(20, pos); pos += 2               // version needed
    centEntry.writeUInt16LE(0, pos); pos += 2                // flags
    centEntry.writeUInt16LE(0, pos); pos += 2                // compression
    centEntry.writeUInt16LE(0, pos); pos += 2                // mod time
    centEntry.writeUInt16LE(0, pos); pos += 2                // mod date
    centEntry.writeUInt32LE(crc, pos); pos += 4              // crc32
    centEntry.writeUInt32LE(data.length, pos); pos += 4      // compressed size
    centEntry.writeUInt32LE(data.length, pos); pos += 4      // uncompressed size
    centEntry.writeUInt16LE(nameBytes.length, pos); pos += 2 // filename length
    centEntry.writeUInt16LE(0, pos); pos += 2                // extra field length
    centEntry.writeUInt16LE(0, pos); pos += 2                // comment length
    centEntry.writeUInt16LE(0, pos); pos += 2                // disk number
    centEntry.writeUInt16LE(0, pos); pos += 2                // internal attrs
    centEntry.writeUInt32LE(0, pos); pos += 4                // external attrs
    centEntry.writeUInt32LE(offset, pos); pos += 4           // local header offset
    Buffer.from(nameBytes).copy(centEntry, pos)

    localHeaders = Buffer.concat([localHeaders, localHeader, data])
    centralDir = Buffer.concat([centralDir, centEntry])
    offset += 30 + nameBytes.length + data.length
  }

  // --- End of Central Directory Record ---
  const eocd = Buffer.alloc(22)
  let pos = 0
  eocd.writeUInt32LE(0x06054b50, pos); pos += 4   // signature
  eocd.writeUInt16LE(0, pos); pos += 2             // disk number
  eocd.writeUInt16LE(0, pos); pos += 2             // central dir disk
  eocd.writeUInt16LE(files.length, pos); pos += 2  // entries on disk
  eocd.writeUInt16LE(files.length, pos); pos += 2  // total entries
  eocd.writeUInt32LE(centralDir.length, pos); pos += 4  // central dir size
  eocd.writeUInt32LE(offset, pos); pos += 4             // central dir offset
  eocd.writeUInt16LE(0, pos); pos += 2                  // comment length

  return Buffer.concat([localHeaders, centralDir, eocd])
}

/**
 * CRC32 计算（用于 ZIP 格式）
 */
function crc32(data) {
  let crc = 0xFFFFFFFF
  for (let i = 0; i < data.length; i++) {
    crc ^= data[i]
    for (let j = 0; j < 8; j++) {
      if (crc & 1) {
        crc = (crc >>> 1) ^ 0xEDB88320
      } else {
        crc = crc >>> 1
      }
    }
  }
  return (crc ^ 0xFFFFFFFF) >>> 0
}

/**
 * 解压 zip 文件到目录
 * @param {string} zipPath - zip 文件路径
 * @param {string} destDir - 目标目录
 */
function unpackZip(zipPath, destDir) {
  const data = fs.readFileSync(zipPath)
  let offset = 0

  fs.mkdirSync(destDir, { recursive: true })

  while (offset < data.length - 4) {
    const sig = data.readUInt32LE(offset)

    // Local File Header signature
    if (sig === 0x04034b50) {
      const nameLen = data.readUInt16LE(offset + 26)
      const extraLen = data.readUInt16LE(offset + 28)
      const compSize = data.readUInt32LE(offset + 18)
      const compression = data.readUInt16LE(offset + 8)

      const nameBuf = data.slice(offset + 30, offset + 30 + nameLen)
      const fileName = new TextDecoder().decode(nameBuf)
      const fileDataStart = offset + 30 + nameLen + extraLen

      if (compression === 0) {
        // Store (no compression)
        const fileData = data.slice(fileDataStart, fileDataStart + compSize)
        const outPath = path.join(destDir, fileName)
        const outDir = path.dirname(outPath)
        fs.mkdirSync(outDir, { recursive: true })
        fs.writeFileSync(outPath, fileData)
      }

      offset = fileDataStart + compSize
    }
    // Central Directory or EOCD → we're done
    else if (sig === 0x02014b50 || sig === 0x06054b50) {
      break
    }
    else {
      offset++
    }
  }
}

module.exports = { packDir, collectFiles, unpackZip }
