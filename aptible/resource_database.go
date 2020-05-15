package aptible

import (
	"log"
	"strconv"

	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatabaseCreate, // POST
		Read:   resourceDatabaseRead,   // GET
		Update: resourceDatabaseUpdate, // PUT
		Delete: resourceDatabaseDelete, // DELETE

		Schema: map[string]*schema.Schema{
			"env_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"handle": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"db_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "postgresql",
				ForceNew: true,
			},
			"container_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1024,
			},
			"disk_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  10,
			},
			"db_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"connection_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	env_id := int64(d.Get("env_id").(int))
	handle := d.Get("handle").(string)
	db_type := d.Get("db_type").(string)
	container_size := int64(d.Get("container_size").(int))
	disk_size := int64(d.Get("disk_size").(int))

	attrs := aptible.DBCreateAttrs{
		Handle:        &handle,
		Type:          db_type,
		ContainerSize: container_size,
		DiskSize:      disk_size,
	}

	db, err := client.CreateDatabase(env_id, attrs)
	if err != nil {
		log.Println(err)
		return err
	}

	d.Set("db_id", db.ID)
	d.Set("connection_url", db.ConnectionURL)
	d.SetId(handle)
	return resourceDatabaseRead(d, meta)
}

// syncs Terraform state with changes made via the API outside of Terraform
func resourceDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	db_id := int64(d.Get("db_id").(int))
	db, deleted, err := client.GetDatabase(db_id)
	if deleted {
		d.SetId("")
		log.Println("Database with ID: " + strconv.Itoa(int(db_id)) + " was deleted outside of Terraform. Now removing it from Terraform state.")
		return nil
	}
	if err != nil {
		log.Println(err)
		return err
	}

	d.Set("container_size", db.ContainerSize)
	d.Set("disk_size", db.DiskSize)
	return nil
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	db_id := int64(d.Get("db_id").(int))
	container_size := int64(d.Get("container_size").(int))
	disk_size := int64(d.Get("disk_size").(int))

	updates := aptible.DBUpdates{}

	if d.HasChange("container_size") {
		updates.ContainerSize = container_size
	}

	if d.HasChange("disk_size") {
		updates.DiskSize = disk_size
	}

	err := client.UpdateDatabase(db_id, updates)
	if err != nil {
		return err
	}

	return resourceDatabaseRead(d, meta)
}

func resourceDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	db_id := int64(d.Get("db_id").(int))
	err := client.DeleteDatabase(db_id)
	if err != nil {
		log.Println(err)
		return err
	}

	d.SetId("")
	return nil
}