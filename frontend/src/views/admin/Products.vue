<template>
  <div class="page-container">
    <div class="page-header">
      <h2>{{ $t('admin.products') }}</h2>
      <div class="header-btns">
        <van-button size="small" icon="description" @click="showInventoryPreview">{{ $t('product.inventoryPreview') }}</van-button>
        <van-button type="success" size="small" plain icon="upgrade" @click="importVisible = true">{{ $t('product.batchImport') }}</van-button>
        <van-button type="primary" size="small" icon="plus" @click="handleAdd">{{ $t('product.addProduct') }}</van-button>
      </div>
    </div>

    <!-- 搜索和分类过滤 -->
    <div class="search-filter-bar">
      <van-search v-model="adminSearchKeyword" :placeholder="$t('product.searchPlaceholder')" shape="round" style="flex:1;padding:0;" />
      <select v-model="adminFilterCat" class="filter-select">
        <option value="">{{ $t('product.all') }}</option>
        <option v-for="c in adminCategories" :key="c" :value="c">{{ c }}</option>
      </select>
    </div>

    <van-loading v-if="loading" size="24" vertical style="padding: 40px 0; text-align:center" />

    <div v-else class="card-list">
      <div v-for="row in paginatedProducts" :key="row.id" class="product-card" @click="handleEdit(row)">
        <div class="pc-left">
          <img v-if="row.img1" :src="row.img1" class="pc-img" />
          <div v-else class="pc-img pc-img-empty"><van-icon name="goods-o" size="24" color="#c0c4cc" /></div>
        </div>
        <div class="pc-right">
          <div class="pc-title-row">
            <span class="pc-name">{{ row.name }}</span>
            <van-tag :type="row.is_active ? 'success' : 'primary'" size="medium" @click.stop="toggleActive(row, !row.is_active)">
              {{ row.is_active ? $t('product.onSale') : $t('product.offSale') }}
            </van-tag>
          </div>
          <div class="pc-meta">
            <span v-if="row.category" class="pc-cat">{{ row.category }}</span>
            <span v-if="row.unit" class="pc-unit">/ {{ row.unit }}</span>
            <van-tag v-if="row.is_featured" type="warning" size="small" style="margin-left:4px">★</van-tag>
            <van-tag v-if="row.is_discounted" type="danger" size="small" style="margin-left:4px">折</van-tag>
          </div>
          <div v-if="row.expiry_date" class="pc-expiry" :style="isExpiringSoon(row.expiry_date) ? 'color:#f56c6c;font-weight:600' : 'color:#909399'">
            {{ $t('product.expiryDate') }}: {{ formatDate(row.expiry_date) }}
          </div>
          <div class="pc-bottom">
            <span class="pc-price">{{ formatUSD(row.price_usd) }}</span>
            <div class="stock-control" @click.stop>
              <button class="stock-btn" @click="openQuickDelta(row, 'pick')" :disabled="row.stock <= 0">−</button>
              <span class="stock-num" :class="{ 'stock-warn': row.is_low_stock }" @click="openStockDrawer(row)">{{ row.stock }}</span>
              <button class="stock-btn" @click="openQuickDelta(row, 'restock')">+</button>
            </div>
          </div>
        </div>
        <div class="pc-actions" @click.stop>
          <van-button type="danger" size="small" plain @click="handleDelete(row)"><van-icon name="delete-o" /></van-button>
        </div>
      </div>
      <van-empty v-if="!loading && products.length === 0" :description="$t('common.noData')" />
    </div>

    <van-pagination
      v-if="filteredAdminProducts.length > pageSize"
      v-model="currentPage"
      :total-items="filteredAdminProducts.length"
      :items-per-page="pageSize"
      :show-page-size="5"
      style="padding: 16px 0;"
    />

    <!-- 库存快捷编辑弹窗 -->
    <van-popup v-model:show="stockDrawerVisible" position="bottom" round :style="{ minHeight: '50vh' }">
      <van-nav-bar
        :title="$t('admin.quickStockEdit')"
        :left-text="$t('common.cancel')"
        :right-text="$t('common.save')"
        @click-left="stockDrawerVisible = false"
        @click-right="saveStockDrawer"
      />
      <div v-if="stockDrawerProduct" style="padding: 16px;">
        <div class="drawer-product-info">
          <span class="drawer-product-name">{{ stockDrawerProduct.name }}</span>
          <van-tag :type="stockDrawerProduct.is_low_stock ? 'warning' : 'success'" size="medium">
            {{ $t('product.stock') }}: {{ stockDrawerProduct.stock }}
          </van-tag>
        </div>
        <van-cell-group inset>
          <van-cell :title="$t('product.restockPick')">
            <template #value>
              <van-radio-group v-model="stockDrawerForm.deltaMode" direction="horizontal">
                <van-radio name="restock">{{ $t('product.restock') }}</van-radio>
                <van-radio name="pick">{{ $t('product.pick') }}</van-radio>
              </van-radio-group>
            </template>
          </van-cell>
          <van-field :label="$t('product.amount')">
            <template #input>
              <van-stepper v-model="stockDrawerForm.deltaAmount" :min="0" />
            </template>
          </van-field>
          <van-cell :title="$t('product.currentStock') + ' →'">
            <template #value>
              <span :class="{ 'stock-warn': previewStock <= (stockDrawerForm.stock_warning || stockDrawerProduct.stock_warning) }">
                {{ stockDrawerProduct.stock }} → {{ previewStock }}
              </span>
            </template>
          </van-cell>
          <van-field v-model.number="stockDrawerForm.stock_warning" type="number" :label="$t('product.stockWarning')" />
          <van-field v-model.number="stockDrawerForm.price_usd" type="number" :label="$t('product.price')" />
        </van-cell-group>
        <div style="padding: 16px 0 0;">
          <van-button block type="primary" :loading="stockDrawerSaving" @click="saveStockDrawer">{{ $t('common.save') }}</van-button>
        </div>
      </div>
    </van-popup>

    <!-- 快捷补货/取货弹窗 -->
    <van-popup v-model:show="quickDeltaVisible" position="bottom" round :style="{ minHeight: '40vh' }" destroy-on-close>
      <van-nav-bar
        :title="quickDeltaMode === 'restock' ? $t('product.restock') : $t('product.pick')"
        :left-text="$t('common.cancel')"
        :right-text="$t('common.confirm')"
        @click-left="quickDeltaVisible = false"
        @click-right="confirmQuickDelta"
      />
      <div v-if="quickDeltaProduct" style="padding: 16px;">
        <div class="quick-delta-info">
          <span>{{ quickDeltaProduct.name }}</span>
          <van-tag :type="quickDeltaProduct.is_low_stock ? 'warning' : 'success'" size="medium">
            {{ $t('product.stock') }}: {{ quickDeltaProduct.stock }}
          </van-tag>
        </div>
        <div style="display:flex;align-items:center;gap:12px;margin:16px 0;">
          <span class="delta-sign">{{ quickDeltaMode === 'restock' ? '+' : '−' }}</span>
          <van-stepper v-model="quickDeltaAmount" :min="1" :max="quickDeltaMode === 'pick' ? quickDeltaProduct.stock : 99999" style="flex:1" />
        </div>
        <div class="quick-delta-preview">
          {{ quickDeltaProduct.stock }} {{ quickDeltaMode === 'restock' ? '+' : '−' }} {{ quickDeltaAmount }}
          = <b>{{ quickDeltaMode === 'restock' ? quickDeltaProduct.stock + quickDeltaAmount : quickDeltaProduct.stock - quickDeltaAmount }}</b>
        </div>
        <van-button block type="primary" :loading="quickDeltaSaving" style="margin-top:16px" @click="confirmQuickDelta">{{ $t('common.confirm') }}</van-button>
      </div>
    </van-popup>

    <!-- 库存预览弹窗 -->
    <van-popup v-model:show="inventoryPreviewVisible" position="bottom" round :style="{ height: '75vh', overflowY: 'auto' }">
      <van-nav-bar :title="$t('product.inventoryPreview')" :left-text="$t('common.cancel')" @click-left="inventoryPreviewVisible = false" />
      <div style="padding: 12px 16px;">
        <div class="inv-summary">
          <div class="inv-stat"><div class="inv-num">{{ products.length }}</div><div class="inv-label">{{ $t('admin.totalProducts') }}</div></div>
          <div class="inv-stat inv-ok"><div class="inv-num">{{ normalStockCount }}</div><div class="inv-label">{{ $t('product.stockNormal') }}</div></div>
          <div class="inv-stat inv-warn"><div class="inv-num">{{ lowStockCount }}</div><div class="inv-label">{{ $t('product.lowStock') }}</div></div>
          <div class="inv-stat inv-danger"><div class="inv-num">{{ outOfStockCount }}</div><div class="inv-label">{{ $t('product.outOfStock') }}</div></div>
        </div>
        <van-tabs v-model:active="invFilter" style="margin: 12px 0;">
          <van-tab :title="$t('product.all')" name="all" />
          <van-tab :title="$t('product.lowStock')" name="low" />
          <van-tab :title="$t('product.outOfStock')" name="out" />
        </van-tabs>
        <div class="inv-list">
          <div v-for="p in filteredInventory" :key="p.id" class="inv-item">
            <span class="inv-dot" :class="getInvDotClass(p)"></span>
            <span class="inv-name">{{ p.name }}</span>
            <span class="inv-stock" :class="getInvDotClass(p)">{{ p.stock }} / {{ p.stock_warning }}</span>
          </div>
          <van-empty v-if="filteredInventory.length === 0" :description="$t('common.noData')" />
        </div>
      </div>
    </van-popup>

    <!-- 编辑/添加商品弹窗 -->
    <van-popup v-model:show="dialogVisible" position="bottom" round :style="{ height: '95vh', overflowY: 'auto' }" destroy-on-close>
      <van-nav-bar
        :title="isEdit ? $t('product.editProduct') : $t('product.addProduct')"
        :left-text="$t('common.cancel')"
        :right-text="$t('common.confirm')"
        @click-left="dialogVisible = false"
        @click-right="handleSubmit"
      />
      <div style="padding-bottom: 20px;">
        <!-- 顶部快捷区 -->
        <div class="form-top-bar">
          <van-cell-group inset>
            <van-field :label="$t('product.category')">
              <template #input>
                <input v-model="form.category" :list="`cat-list-${Date.now()}`" :placeholder="$t('product.selectCategory')" style="border:none;outline:none;width:100%;font-size:14px;" />
                <datalist id="cat-list">
                  <option v-for="cat in categoryOptions" :key="cat" :value="cat" />
                </datalist>
              </template>
            </van-field>
            <van-field v-model="form.sort_order" type="number" :label="$t('productForm.no')" />
            <van-field v-model="form.barcode" :label="$t('productForm.barcode')" clearable>
              <template #button>
                <van-button size="small" plain icon="scan" @click="openBarcodeScan">{{ $t('product.scanBarcode') }}</van-button>
              </template>
            </van-field>
          </van-cell-group>
          <div class="form-flag-row">
            <div class="flag-item" :class="{ 'flag-active': form.is_active }">
              <span class="flag-label">{{ $t('product.status') }}</span>
              <van-switch v-model="form.is_active" size="20" />
            </div>
            <div class="flag-item" :class="{ 'flag-featured': form.is_featured }">
              <span class="flag-label">{{ $t('product.featured') }}</span>
              <van-switch v-model="form.is_featured" size="20" active-color="#f59e0b" />
            </div>
            <div class="flag-item" :class="{ 'flag-discount': form.is_discounted }">
              <span class="flag-label">{{ $t('product.discounted') }}</span>
              <van-switch v-model="form.is_discounted" size="20" active-color="#ef4444" />
            </div>
          </div>
        </div>

        <van-collapse v-model="openSections">
          <!-- ① 基本信息 -->
          <van-collapse-item :title="$t('productForm.sectionBasic')" name="basic">
            <van-cell-group>
              <van-field v-model="form.name" :label="$t('productForm.nameZh')" :placeholder="$t('productForm.nameZh')" required />
              <van-field v-model="form.name_en" :label="$t('productForm.nameEn')" />
              <van-field v-model="form.name_kh" :label="$t('productForm.nameKh')" />
              <van-field v-model="form.unit" :label="$t('productForm.unitLabel')" :placeholder="$t('product.unitPlaceholder')" />
              <van-field v-model="form.price_usd" type="number" :label="$t('productForm.costUSD')" placeholder="0.00">
                <template #extra><span style="color:#999;font-size:12px;">USD</span></template>
              </van-field>
              <van-field v-model="form.stock" type="number" :label="$t('product.stock')" placeholder="0" />
              <van-field v-model="form.stock_warning" type="number" :label="$t('product.stockWarning')" placeholder="10" />
              <van-field v-model="form.description" type="textarea" rows="2" :label="$t('product.description')" />
            </van-cell-group>

            <!-- 图片上传 -->
            <div style="padding: 12px 16px 0;">
              <div style="font-size:13px;color:#909399;margin-bottom:8px;">{{ $t('product.image') }}</div>
              <div class="image-upload-area" @paste="onPasteImage">
                <div v-for="idx in 5" :key="idx" class="image-slot">
                  <div v-if="form[`img${idx}`]" class="image-preview">
                    <img :src="form[`img${idx}`]" style="width:100%;height:100%;object-fit:cover;" />
                    <div class="image-actions" @click="removeImage(idx)">
                      <van-icon name="delete-o" size="22" color="#fff" />
                    </div>
                  </div>
                  <div v-else class="upload-placeholder" @click="triggerImageUpload(idx)">
                    <van-icon name="plus" size="20" color="#999" />
                    <span>{{ idx }}</span>
                  </div>
                  <input :ref="el => { if (el) imageInputRefs[idx] = el }" type="file" accept="image/*" style="display:none" @change="e => handleFileChange(e, idx)" />
                </div>
              </div>
              <div class="upload-tip">{{ $t('product.uploadTip5') }} | {{ $t('product.pasteTip') }}</div>
            </div>
          </van-collapse-item>

          <!-- ② 其他信息 -->
          <van-collapse-item :title="$t('productForm.sectionMoreInfo')" name="moreinfo">
            <van-cell-group>
              <van-field v-model="form.brand" :label="$t('productForm.brand')" />
              <van-field v-model="form.country_of_origin" :label="$t('productForm.countryOfOrigin')" />
              <van-field v-model="form.supplier_name" :label="$t('productForm.supplierName')" />
              <van-field v-model="form.principle_company" :label="$t('productForm.principleCompany')" />
              <van-field v-model="form.packing_format" :label="$t('productForm.packingFormat')" :placeholder="$t('productForm.packingFormatPlaceholder')" />
              <van-field :label="$t('productForm.unitWeightValue')">
                <template #input>
                  <input v-model="form.unit_weight_value" type="number" placeholder="0.00" style="flex:1;border:none;outline:none;font-size:14px;" />
                  <select v-model="form.unit_weight_unit" style="border:none;border-left:1px solid #ebedf0;padding:0 8px;background:transparent;font-size:14px;">
                    <option value="G">G</option>
                    <option value="ML">ML</option>
                    <option value="Pcs">Pcs</option>
                  </select>
                </template>
              </van-field>
              <van-field v-model="form.pack_size" type="number" :label="$t('productForm.packSize')" />
              <van-field v-model="form.shelf_life_days" type="number" :label="$t('productForm.shelfLifeDays')" />
              <van-field :label="$t('product.productionDate')">
                <template #input>
                  <input type="date" v-model="form.production_date" style="border:none;outline:none;width:100%;font-size:14px;" />
                </template>
              </van-field>
              <van-field :label="$t('product.expiryDate')">
                <template #input>
                  <input type="date" v-model="form.expiry_date" style="border:none;outline:none;width:100%;font-size:14px;" />
                </template>
              </van-field>
            </van-cell-group>
          </van-collapse-item>

          <!-- ③ 包装规格 -->
          <van-collapse-item :title="$t('productForm.sectionPacking')" name="packing">
            <van-cell-group>
              <div class="dim-group-title">{{ $t('productForm.singlePackTitle') }}</div>
              <van-field v-model="form.unit_name" :label="$t('productForm.unitLabel')" :placeholder="$t('product.unitNamePlaceholder')" />
              <van-field v-model="form.price_per_piece_usd" type="number" :label="$t('productForm.pricePerPiece')" placeholder="0.00" />
              <div class="dim-group-title">{{ $t('productForm.bulkPackTitle') }}</div>
              <van-field v-model="form.pack_name" :label="$t('productForm.packLabel')" :placeholder="$t('product.packNamePlaceholder')" />
              <van-field v-model="form.unit_per_inner_pack" type="number" :label="$t('productForm.unitPerInnerPack')" placeholder="0" />
              <van-field v-model="form.price_per_package_usd" type="number" :label="$t('productForm.pricePerPackage')" placeholder="0.00" />
              <div class="dim-group-title">{{ $t('productForm.caseTitle') }}</div>
              <van-field v-model="form.inner_pack_per_case" type="number" :label="$t('productForm.innerPackPerCase')" placeholder="0" />
              <van-field v-model="form.unit_per_case" type="number" :label="$t('productForm.unitPerCase')" placeholder="0" />
              <van-field v-model="form.price_per_case_usd" type="number" :label="$t('productForm.pricePerCase')" placeholder="0.00" />
            </van-cell-group>
          </van-collapse-item>

          <!-- ④ 尺寸与重量 -->
          <van-collapse-item :title="$t('productForm.sectionDimensions')" name="dimensions">
            <van-cell-group>
              <div class="dim-group-title">{{ $t('productForm.smallestUnit') }}</div>
              <van-field v-model="form.unit_width_cm" type="number" :label="$t('productForm.widthCm')" />
              <van-field v-model="form.unit_length_cm" type="number" :label="$t('productForm.lengthCm')" />
              <van-field v-model="form.unit_height_cm" type="number" :label="$t('productForm.heightCm')" />
              <van-field v-model="form.unit_weight_kg" type="number" :label="$t('productForm.weightKg')" />
              <div class="dim-group-title">{{ $t('productForm.middlePack') }}</div>
              <van-field v-model="form.pack_width_cm" type="number" :label="$t('productForm.widthCm')" />
              <van-field v-model="form.pack_length_cm" type="number" :label="$t('productForm.lengthCm')" />
              <van-field v-model="form.pack_height_cm" type="number" :label="$t('productForm.heightCm')" />
              <van-field v-model="form.pack_weight_kg" type="number" :label="$t('productForm.weightKg')" />
              <div class="dim-group-title">{{ $t('productForm.outerCase') }}</div>
              <van-field v-model="form.case_width_cm" type="number" :label="$t('productForm.widthCm')" />
              <van-field v-model="form.case_length_cm" type="number" :label="$t('productForm.lengthCm')" />
              <van-field v-model="form.case_height_cm" type="number" :label="$t('productForm.heightCm')" />
              <van-field v-model="form.case_weight_kg" type="number" :label="$t('productForm.weightKg')" />
            </van-cell-group>
          </van-collapse-item>
        </van-collapse>

        <div style="padding: 16px;">
          <van-button block type="primary" :loading="submitting" @click="handleSubmit">{{ $t('common.confirm') }}</van-button>
        </div>
      </div>
    </van-popup>

    <!-- 批量导入弹窗 -->
    <van-popup v-model:show="importVisible" position="bottom" round :style="{ minHeight: '55vh' }">
      <van-nav-bar
        :title="$t('product.batchImport')"
        :left-text="$t('common.cancel')"
        @click-left="importVisible = false"
      />
      <div style="padding: 16px;">
        <p style="font-size:13px;color:#606266;margin-bottom:12px;">{{ $t('product.batchImportTip') }}</p>
        <a :href="templateUrl" download style="color:#1989fa;font-size:13px;text-decoration:underline;">{{ $t('product.downloadTemplate') }}</a>
        <div style="margin-top: 16px; display:flex; align-items:center; gap:12px;">
          <van-button type="primary" size="small" @click="importInputRef?.click()">{{ $t('product.selectFile') }}</van-button>
          <span v-if="importFileName" style="font-size:13px;color:#666;">{{ importFileName }}</span>
        </div>
        <input ref="importInputRef" type="file" accept=".csv,.xlsx" style="display:none" @change="onImportFileChange" />
        <van-checkbox v-model="importOverwrite" style="margin-top: 12px;">{{ $t('product.overwriteExisting') }}</van-checkbox>
        <div v-if="importResult" class="import-result">
          <p>{{ $t('product.imported') }}: {{ importResult.created }} / {{ $t('product.updated') }}: {{ importResult.updated }} / {{ $t('product.skipped') }}: {{ importResult.skipped }}</p>
          <ul v-if="importResult.errors?.length" class="import-errors">
            <li v-for="(err, i) in importResult.errors" :key="i">{{ err }}</li>
          </ul>
        </div>
        <van-button block type="primary" :loading="importing" :disabled="!importFile" style="margin-top:16px" @click="submitImport">
          {{ $t('product.startImport') }}
        </van-button>
      </div>
    </van-popup>

    <!-- 扫码弹窗 -->
    <van-popup v-model:show="scanVisible" position="bottom" round :style="{ height: '65vh' }" @close="stopScan">
      <van-nav-bar
        :title="$t('product.scanBarcode')"
        :left-text="$t('common.cancel')"
        @click-left="scanVisible = false; stopScan()"
      />
      <div style="padding: 0 16px 16px;">
        <div v-if="scanError" class="scan-error">{{ scanError }}</div>
        <video v-show="!scanError" ref="scanVideoRef" class="scan-video" autoplay muted playsinline></video>
        <div style="margin-top: 12px;">
          <van-field v-model="manualBarcode" :placeholder="$t('product.enterBarcode')" clearable>
            <template #button>
              <van-button size="small" type="primary" @click="confirmManual">{{ $t('common.confirm') }}</van-button>
            </template>
          </van-field>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { showSuccessToast, showFailToast, showConfirmDialog } from 'vant'
import { getProducts, createProduct, updateProduct, deleteProduct, uploadImage, getAllCategories, importProducts, getProductImportTemplateUrl } from '@/api'
import { formatUSD } from '@/utils/format'

const formatDate = (d) => {
  if (!d) return ''
  return new Date(d).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}
const isExpiringSoon = (d, days = 30) => {
  if (!d) return false
  return (new Date(d) - Date.now()) / 86400000 <= days
}

const { t } = useI18n()

const loading = ref(false)
const products = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const categoryOptions = ref([])

// 批量导入
const importVisible = ref(false)
const importFile = ref(null)
const importFileName = ref('')
const importOverwrite = ref(false)
const importing = ref(false)
const importResult = ref(null)
const importInputRef = ref()
const templateUrl = getProductImportTemplateUrl()

// 扫码
const scanVisible = ref(false)
const scanError = ref('')
const scanVideoRef = ref()
const manualBarcode = ref('')
let scanStream = null
let scanDetector = null
let scanRAF = 0

// 图片上传 refs
const imageInputRefs = {}

const triggerImageUpload = (idx) => {
  if (imageInputRefs[idx]) imageInputRefs[idx].click()
}

const handleFileChange = async (event, idx) => {
  const file = event.target.files?.[0]
  if (!file) return
  try {
    const res = await uploadImage(file)
    form[`img${idx}`] = res.url
    if (idx === 1) form.image_url = res.url
    showSuccessToast(`${t('product.image')}${idx} ✓`)
  } catch {
    showFailToast(t('product.uploadFailed'))
  }
  event.target.value = ''
}

const onImportFileChange = (event) => {
  const file = event.target.files?.[0]
  importFile.value = file || null
  importFileName.value = file?.name || ''
  importResult.value = null
}

const submitImport = async () => {
  if (!importFile.value) return
  importing.value = true
  importResult.value = null
  try {
    const fd = new FormData()
    fd.append('file', importFile.value)
    const data = await importProducts(fd)
    importResult.value = data
    showSuccessToast(t('product.importDone'))
    await loadProducts()
  } catch (e) {
    console.error(e)
  } finally {
    importing.value = false
  }
}

const stopScan = () => {
  if (scanRAF) cancelAnimationFrame(scanRAF)
  scanRAF = 0
  scanDetector = null
  if (scanStream) { scanStream.getTracks().forEach(t => t.stop()); scanStream = null }
}

const openBarcodeScan = async () => {
  scanError.value = ''
  manualBarcode.value = ''
  scanVisible.value = true
  await nextTick()
  if (!('BarcodeDetector' in window)) { scanError.value = t('product.scanNotSupported'); return }
  try {
    scanStream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: 'environment' } })
    if (scanVideoRef.value) scanVideoRef.value.srcObject = scanStream
    scanDetector = new window.BarcodeDetector({ formats: ['ean_13', 'ean_8', 'code_128', 'code_39', 'upc_a', 'upc_e', 'qr_code'] })
    const tick = async () => {
      if (!scanDetector || !scanVideoRef.value) return
      try {
        const codes = await scanDetector.detect(scanVideoRef.value)
        if (codes && codes.length > 0) {
          form.barcode = codes[0].rawValue
          showSuccessToast(t('product.scanSuccess'))
          scanVisible.value = false
          stopScan()
          return
        }
      } catch {}
      scanRAF = requestAnimationFrame(tick)
    }
    scanRAF = requestAnimationFrame(tick)
  } catch { scanError.value = t('product.scanCameraError') }
}

const confirmManual = () => {
  if (!manualBarcode.value) return
  form.barcode = manualBarcode.value.trim()
  scanVisible.value = false
  stopScan()
}

// 分页
const currentPage = ref(1)
const pageSize = ref(20)
const adminSearchKeyword = ref('')
const adminFilterCat = ref('')

const adminCategories = computed(() => {
  const cats = [...new Set(products.value.map(p => p.category).filter(Boolean))]
  return cats.sort()
})

const filteredAdminProducts = computed(() => {
  let list = products.value
  if (adminFilterCat.value) list = list.filter(p => p.category === adminFilterCat.value)
  if (adminSearchKeyword.value.trim()) {
    const kw = adminSearchKeyword.value.toLowerCase()
    list = list.filter(p =>
      (p.name && p.name.toLowerCase().includes(kw)) ||
      (p.name_kh && p.name_kh.toLowerCase().includes(kw)) ||
      (p.name_en && p.name_en.toLowerCase().includes(kw)) ||
      (p.category && p.category.toLowerCase().includes(kw)) ||
      (p.brand && p.brand.toLowerCase().includes(kw))
    )
  }
  return list
})

const paginatedProducts = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return filteredAdminProducts.value.slice(start, start + pageSize.value)
})

const form = reactive({
  id: null,
  supplier_name: '', sort_order: 0, barcode: '', brand: '', category: '', name: '',
  unit_weight_value: '', unit_weight_unit: 'G', packing_format: '', pack_size: '',
  price_usd: '', gp_percent: '', shelf_life_days: '', principle_company: '',
  country_of_origin: '', production_date: null, expiry_date: null,
  image_url: '', img1: '', img2: '', img3: '', img4: '', img5: '',
  stock: 0, stock_warning: 10, is_active: true, is_featured: false, is_discounted: false,
  name_en: '', name_kh: '', description: '',
  unit: '', pack_name: '', unit_name: '',
  inner_pack_per_case: '', unit_per_inner_pack: '', unit_per_case: '', pieces_per_package: '',
  cost_per_case: '', dc_percent: '', net_cost_per_case: '', net_cost_per_unit: '',
  price_incl_vat: '', price_excl_vat: '',
  price_per_piece_usd: '', price_per_package_usd: '', price_per_case_usd: '',
  unit_width_cm: '', unit_length_cm: '', unit_height_cm: '', unit_weight_kg: '',
  pack_width_cm: '', pack_length_cm: '', pack_height_cm: '', pack_weight_kg: '',
  case_width_cm: '', case_length_cm: '', case_height_cm: '', case_weight_kg: '',
  specs: '',
})

const openSections = ref(['basic', 'packing'])

const getImageList = (row) => [row.img1, row.img2, row.img3, row.img4, row.img5].filter(Boolean)

const removeImage = (idx) => {
  form[`img${idx}`] = ''
  if (idx === 1) form.image_url = ''
}

const onPasteImage = async (e) => {
  const items = e.clipboardData?.items
  if (!items) return
  for (const item of items) {
    if (item.type.startsWith('image/')) {
      e.preventDefault()
      const file = item.getAsFile()
      if (!file) return
      const emptyIdx = [1, 2, 3, 4, 5].find(i => !form[`img${i}`])
      if (!emptyIdx) { showFailToast(t('product.uploadTip')); return }
      await handleFileChange({ target: { files: [file], value: '' } }, emptyIdx)
      break
    }
  }
}

const loadCategories = async () => {
  try {
    const data = await getAllCategories()
    categoryOptions.value = data.map(c => c.name)
  } catch {
    categoryOptions.value = [...new Set(products.value.map(p => p.category).filter(Boolean))]
  }
}

const loadProducts = async () => {
  loading.value = true
  try {
    products.value = await getProducts({ is_active: null })
  } catch (error) {
    console.error('加载商品失败:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  Object.assign(form, {
    id: null, supplier_name: '', sort_order: 0, barcode: '', brand: '', category: '', name: '',
    unit_weight_value: '', unit_weight_unit: 'G', packing_format: '', pack_size: '',
    price_usd: '', gp_percent: '', shelf_life_days: '', principle_company: '', country_of_origin: '',
    production_date: null, expiry_date: null,
    image_url: '', img1: '', img2: '', img3: '', img4: '', img5: '',
    stock: 0, stock_warning: 10, is_active: true, is_featured: false, is_discounted: false,
    name_en: '', name_kh: '', description: '',
    unit: '', pack_name: '', unit_name: '',
    inner_pack_per_case: '', unit_per_inner_pack: '', unit_per_case: '', pieces_per_package: '',
    cost_per_case: '', dc_percent: '', net_cost_per_case: '', net_cost_per_unit: '',
    price_incl_vat: '', price_excl_vat: '',
    price_per_piece_usd: '', price_per_package_usd: '', price_per_case_usd: '',
    unit_width_cm: '', unit_length_cm: '', unit_height_cm: '', unit_weight_kg: '',
    pack_width_cm: '', pack_length_cm: '', pack_height_cm: '', pack_weight_kg: '',
    case_width_cm: '', case_length_cm: '', case_height_cm: '', case_weight_kg: '',
    specs: '',
  })
}

watch(
  () => [form.production_date, form.shelf_life_days],
  ([prodDate, days]) => {
    if (prodDate && days && Number(days) > 0) {
      const d = new Date(prodDate)
      d.setDate(d.getDate() + Number(days))
      form.expiry_date = d.toISOString().substring(0, 10)
    }
  }
)

const toStr = (v) => (v != null ? String(v) : '')

const handleAdd = () => { resetForm(); isEdit.value = false; dialogVisible.value = true }

const handleEdit = (row) => {
  Object.assign(form, {
    ...row,
    supplier_name: row.supplier_name || '', name: row.name || '',
    name_en: row.name_en || '', name_kh: row.name_kh || '',
    brand: row.brand || '', country_of_origin: row.country_of_origin || '',
    packing_format: row.packing_format || '', principle_company: row.principle_company || '',
    unit: row.unit || '', unit_name: row.unit_name || '', pack_name: row.pack_name || '',
    barcode: row.barcode || '', description: row.description || '', specs: row.specs || '',
    unit_weight_unit: row.unit_weight_unit || 'G',
    img1: row.img1 || '', img2: row.img2 || '', img3: row.img3 || '',
    img4: row.img4 || '', img5: row.img5 || '',
    production_date: row.production_date ? row.production_date.substring(0, 10) : null,
    expiry_date: row.expiry_date ? row.expiry_date.substring(0, 10) : null,
    sort_order: row.sort_order || 0,
    is_active: row.is_active !== false, is_featured: row.is_featured || false, is_discounted: row.is_discounted || false,
    price_usd: toStr(row.price_usd), price_per_piece_usd: toStr(row.price_per_piece_usd),
    price_per_package_usd: toStr(row.price_per_package_usd), price_per_case_usd: toStr(row.price_per_case_usd),
    unit_per_inner_pack: toStr(row.unit_per_inner_pack || row.pieces_per_package),
    pieces_per_package: toStr(row.unit_per_inner_pack || row.pieces_per_package),
    unit_weight_value: toStr(row.unit_weight_value), pack_size: toStr(row.pack_size),
    gp_percent: toStr(row.gp_percent), shelf_life_days: toStr(row.shelf_life_days),
    inner_pack_per_case: toStr(row.inner_pack_per_case), unit_per_case: toStr(row.unit_per_case),
    cost_per_case: toStr(row.cost_per_case), dc_percent: toStr(row.dc_percent),
    net_cost_per_case: toStr(row.net_cost_per_case), net_cost_per_unit: toStr(row.net_cost_per_unit),
    price_incl_vat: toStr(row.price_incl_vat), price_excl_vat: toStr(row.price_excl_vat),
    unit_width_cm: toStr(row.unit_width_cm), unit_length_cm: toStr(row.unit_length_cm),
    unit_height_cm: toStr(row.unit_height_cm), unit_weight_kg: toStr(row.unit_weight_kg),
    pack_width_cm: toStr(row.pack_width_cm), pack_length_cm: toStr(row.pack_length_cm),
    pack_height_cm: toStr(row.pack_height_cm), pack_weight_kg: toStr(row.pack_weight_kg),
    case_width_cm: toStr(row.case_width_cm), case_length_cm: toStr(row.case_length_cm),
    case_height_cm: toStr(row.case_height_cm), case_weight_kg: toStr(row.case_weight_kg),
  })
  isEdit.value = true
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!form.name?.trim()) { showFailToast(t('product.nameRequired')); return }
  const priceVal = parseFloat(form.price_usd)
  if (isNaN(priceVal) || priceVal <= 0) { showFailToast(t('product.priceRequired')); return }
  submitting.value = true
  try {
    const payload = { ...form }
    delete payload.id
    if (payload.unit_per_inner_pack) payload.pieces_per_package = payload.unit_per_inner_pack
    const floatFields = [
      'price_usd', 'price_per_piece_usd', 'price_per_package_usd',
      'unit_weight_value', 'pack_size', 'gp_percent',
      'cost_per_case', 'dc_percent', 'net_cost_per_case', 'net_cost_per_unit', 'price_incl_vat', 'price_excl_vat',
      'unit_width_cm', 'unit_length_cm', 'unit_height_cm', 'unit_weight_kg',
      'pack_width_cm', 'pack_length_cm', 'pack_height_cm', 'pack_weight_kg',
      'case_width_cm', 'case_length_cm', 'case_height_cm', 'case_weight_kg',
    ]
    const intFields = [
      'pieces_per_package', 'stock', 'stock_warning', 'sort_order',
      'shelf_life_days', 'inner_pack_per_case', 'unit_per_inner_pack', 'unit_per_case',
    ]
    for (const key of floatFields) { const n = parseFloat(payload[key]); payload[key] = isNaN(n) ? null : n }
    for (const key of intFields) { const n = parseInt(payload[key]); payload[key] = isNaN(n) ? null : n }
    for (const key of Object.keys(payload)) { if (payload[key] === '') payload[key] = null }
    if (isEdit.value) {
      await updateProduct(form.id, payload)
      showSuccessToast(t('product.productUpdated'))
    } else {
      delete payload.is_active
      await createProduct(payload)
      showSuccessToast(t('product.productAdded'))
    }
    dialogVisible.value = false
    loadProducts()
  } catch (error) {
    console.error('提交失败:', error)
  } finally {
    submitting.value = false
  }
}

const toggleActive = async (row, val) => {
  try {
    await updateProduct(row.id, { is_active: val })
    row.is_active = val
    showSuccessToast(val ? t('product.statusOnSale') : t('product.statusOffSale'))
  } catch { showFailToast(t('common.operationFailed')) }
}

const toggleFeatured = async (row, val) => {
  try {
    await updateProduct(row.id, { is_featured: val })
    row.is_featured = val
  } catch {}
}

const toggleDiscounted = async (row, val) => {
  try {
    await updateProduct(row.id, { is_discounted: val })
    row.is_discounted = val
  } catch {}
}

// 库存快捷操作
const stockDrawerVisible = ref(false)
const stockDrawerProduct = ref(null)
const stockDrawerSaving = ref(false)
const stockDrawerForm = reactive({ deltaMode: 'restock', deltaAmount: 0, stock_warning: 10, price_usd: 0 })

const previewStock = computed(() => {
  if (!stockDrawerProduct.value) return 0
  const base = stockDrawerProduct.value.stock
  return stockDrawerForm.deltaMode === 'restock'
    ? base + stockDrawerForm.deltaAmount
    : Math.max(0, base - stockDrawerForm.deltaAmount)
})

const quickDeltaVisible = ref(false)
const quickDeltaProduct = ref(null)
const quickDeltaMode = ref('restock')
const quickDeltaAmount = ref(1)
const quickDeltaSaving = ref(false)

const inventoryPreviewVisible = ref(false)
const invFilter = ref('all')

const lowStockCount = computed(() => products.value.filter(p => p.stock > 0 && p.is_low_stock).length)
const outOfStockCount = computed(() => products.value.filter(p => p.stock <= 0).length)
const normalStockCount = computed(() => products.value.filter(p => !p.is_low_stock && p.stock > 0).length)

const filteredInventory = computed(() => {
  if (invFilter.value === 'low') return products.value.filter(p => p.stock > 0 && p.is_low_stock)
  if (invFilter.value === 'out') return products.value.filter(p => p.stock <= 0)
  return [...products.value].sort((a, b) => a.stock - b.stock)
})

const getInvDotClass = (p) => {
  if (p.stock <= 0) return 'inv-danger'
  if (p.is_low_stock) return 'inv-warn'
  return 'inv-ok'
}

const showInventoryPreview = () => { invFilter.value = 'all'; inventoryPreviewVisible.value = true }

const openQuickDelta = (row, mode) => {
  quickDeltaProduct.value = row
  quickDeltaMode.value = mode
  quickDeltaAmount.value = 1
  quickDeltaVisible.value = true
}

const confirmQuickDelta = async () => {
  if (!quickDeltaProduct.value) return
  const row = quickDeltaProduct.value
  const delta = quickDeltaMode.value === 'restock' ? quickDeltaAmount.value : -quickDeltaAmount.value
  const newStock = Math.max(0, row.stock + delta)
  quickDeltaSaving.value = true
  try {
    await updateProduct(row.id, { stock: newStock })
    row.stock = newStock
    row.is_low_stock = newStock <= (row.stock_warning || 0)
    showSuccessToast(t('product.productUpdated'))
    quickDeltaVisible.value = false
  } catch { showFailToast(t('common.operationFailed')) }
  finally { quickDeltaSaving.value = false }
}

const openStockDrawer = (row) => {
  stockDrawerProduct.value = row
  Object.assign(stockDrawerForm, { deltaMode: 'restock', deltaAmount: 0, stock_warning: row.stock_warning, price_usd: row.price_usd })
  stockDrawerVisible.value = true
}

const saveStockDrawer = async () => {
  if (!stockDrawerProduct.value) return
  stockDrawerSaving.value = true
  try {
    const newStock = previewStock.value
    await updateProduct(stockDrawerProduct.value.id, {
      stock: newStock, stock_warning: stockDrawerForm.stock_warning, price_usd: stockDrawerForm.price_usd,
    })
    stockDrawerProduct.value.stock = newStock
    stockDrawerProduct.value.stock_warning = stockDrawerForm.stock_warning
    stockDrawerProduct.value.price_usd = stockDrawerForm.price_usd
    stockDrawerProduct.value.is_low_stock = newStock <= stockDrawerForm.stock_warning
    showSuccessToast(t('product.productUpdated'))
    stockDrawerVisible.value = false
    loadProducts()
  } catch (error) {
    console.error('保存失败:', error)
  } finally {
    stockDrawerSaving.value = false
  }
}

const handleDelete = async (row) => {
  try {
    await showConfirmDialog({
      title: t('product.deleteTitle'),
      message: t('product.deleteConfirm', { name: row.name }),
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
    })
    await deleteProduct(row.id)
    showSuccessToast(t('product.productDeleted'))
    loadProducts()
  } catch {}
}

onMounted(async () => {
  await loadProducts()
  loadCategories()
})
onBeforeUnmount(() => stopScan())
</script>

<style scoped>
.header-btns { display: flex; gap: 6px; flex-wrap: wrap; }

.search-filter-bar {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 12px;
}

.filter-select {
  height: 34px;
  padding: 0 8px;
  border: 1px solid #ebedf0;
  border-radius: 4px;
  font-size: 14px;
  background: #fff;
  color: #323233;
  flex-shrink: 0;
  width: 120px;
}

.card-list { display: flex; flex-direction: column; gap: 10px; }

.product-card {
  display: flex;
  gap: 12px;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 10px;
  padding: 12px;
  cursor: pointer;
  align-items: center;
}
.product-card:active { background: #f5f7fa; }

.pc-left { flex-shrink: 0; }
.pc-img { width: 72px; height: 72px; border-radius: 8px; object-fit: cover; }
.pc-img-empty { display: flex; align-items: center; justify-content: center; background: #f5f7fa; }
.pc-right { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4px; }
.pc-title-row { display: flex; align-items: center; gap: 6px; overflow: hidden; }
.pc-name { font-size: 15px; font-weight: 600; color: #303133; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1; min-width: 0; }
.pc-meta { font-size: 12px; color: #909399; display: flex; align-items: center; flex-wrap: wrap; gap: 2px; }
.pc-cat { background: #f0f2f5; padding: 1px 6px; border-radius: 3px; }
.pc-expiry { font-size: 12px; }
.pc-bottom { display: flex; justify-content: space-between; align-items: center; }
.pc-price { font-size: 15px; font-weight: 700; color: #1D4ED8; }
.pc-actions { flex-shrink: 0; }

.stock-control { display: flex; align-items: center; background: #f5f5f5; border-radius: 6px; border: 1px solid #e4e7ed; overflow: hidden; }
.stock-btn { width: 28px; height: 28px; border: none; background: transparent; color: #606266; font-size: 16px; font-weight: 700; cursor: pointer; display: flex; align-items: center; justify-content: center; }
.stock-btn:hover { background: #e8e8e8; }
.stock-btn:disabled { color: #c0c4cc; cursor: not-allowed; }
.stock-num { min-width: 34px; text-align: center; font-size: 14px; font-weight: 600; color: #303133; cursor: pointer; padding: 0 4px; }
.stock-num.stock-warn { color: #E6A23C; }

.form-top-bar { margin-bottom: 0; }
.form-flag-row { display: flex; gap: 8px; padding: 12px 16px; flex-wrap: wrap; }
.flag-item { display: flex; align-items: center; gap: 8px; background: #fff; border: 1px solid #e4e7ed; border-radius: 6px; padding: 6px 12px; flex: 1; min-width: 0; }
.flag-item.flag-active { border-color: #67c23a; background: #f0fdf4; }
.flag-item.flag-featured { border-color: #f59e0b; background: #fffbeb; }
.flag-item.flag-discount { border-color: #ef4444; background: #fff5f5; }
.flag-label { font-size: 12px; color: #606266; white-space: nowrap; flex: 1; }

.dim-group-title { font-size: 13px; font-weight: 600; color: #555; margin: 8px 16px 0; padding-left: 6px; border-left: 3px solid #1989fa; }

.image-upload-area { display: flex; gap: 10px; flex-wrap: wrap; }
.image-slot { width: 80px; height: 80px; border-radius: 6px; overflow: hidden; }
.image-preview { position: relative; width: 100%; height: 100%; border: 1px solid #eee; border-radius: 6px; overflow: hidden; }
.image-actions { position: absolute; top: 0; left: 0; width: 100%; height: 100%; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.4); opacity: 0; transition: opacity 0.2s; cursor: pointer; }
.image-preview:hover .image-actions { opacity: 1; }
.upload-placeholder { width: 100%; height: 100%; border: 1px dashed #d9d9d9; border-radius: 6px; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 4px; color: #999; font-size: 12px; cursor: pointer; }
.upload-placeholder:hover { border-color: #1989fa; }
.upload-tip { font-size: 12px; color: #999; margin-top: 6px; }

.drawer-product-info { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.drawer-product-name { font-size: 16px; font-weight: 600; color: #303133; }

.quick-delta-info { display: flex; justify-content: space-between; align-items: center; font-weight: 600; }
.delta-sign { font-size: 24px; font-weight: 700; color: #1989fa; width: 28px; text-align: center; }
.quick-delta-preview { text-align: center; font-size: 14px; color: #909399; padding: 8px; background: #f5f7fa; border-radius: 6px; }

.inv-summary { display: flex; justify-content: space-around; gap: 10px; margin-bottom: 8px; }
.inv-stat { text-align: center; flex: 1; padding: 10px 0; }
.inv-num { font-size: 24px; font-weight: 700; color: #303133; }
.inv-label { font-size: 12px; color: #909399; margin-top: 4px; }
.inv-ok .inv-num { color: #67c23a; }
.inv-warn .inv-num { color: #E6A23C; }
.inv-danger .inv-num { color: #f56c6c; }
.inv-list { display: flex; flex-direction: column; gap: 6px; max-height: 280px; overflow-y: auto; }
.inv-item { display: flex; align-items: center; gap: 8px; padding: 8px 10px; border-radius: 6px; background: #fafafa; }
.inv-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.inv-dot.inv-ok { background: #67c23a; }
.inv-dot.inv-warn { background: #E6A23C; }
.inv-dot.inv-danger { background: #f56c6c; }
.inv-name { flex: 1; font-size: 14px; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.inv-stock { font-weight: 600; font-size: 13px; flex-shrink: 0; }
.inv-stock.inv-ok { color: #67c23a; }
.inv-stock.inv-warn { color: #E6A23C; }
.inv-stock.inv-danger { color: #f56c6c; }

.import-result { margin-top: 12px; padding: 8px 12px; background: #f5f7fa; border-radius: 4px; font-size: 13px; }
.import-errors { margin: 6px 0 0 16px; color: #f56c6c; max-height: 160px; overflow-y: auto; }

.scan-video { width: 100%; max-height: 50vh; background: #000; border-radius: 4px; margin-top: 8px; }
.scan-error { padding: 12px; color: #f56c6c; background: #fef0f0; border-radius: 4px; margin-bottom: 12px; }
</style>
